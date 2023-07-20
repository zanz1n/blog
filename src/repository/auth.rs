use super::cache::{CacheRepository, CacheService};
use crate::error::ApiError;
use crate::model::user::{Column as UserColumn, Entity as UserEntity, UserRole};
use crate::utils::generic::now_unix_sec;
use crate::utils::http::DataBody;
use actix_web::body::BoxBody;
use actix_web::Responder;
use async_trait::async_trait;
use jsonwebtoken::{
    errors::ErrorKind as JwtErrorKind, Algorithm, DecodingKey, EncodingKey, Header, TokenData,
    Validation,
};
use sea_orm::DatabaseConnection;
use sea_orm::{ColumnTrait, EntityTrait, QueryFilter, QuerySelect};
use serde::{Deserialize, Serialize};
use tokio::task::spawn_blocking;

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct UserJwtPayload {
    pub sub: String,
    pub username: String,
    pub email: String,
    pub exp: u64,
    pub iat: u64,
    pub role: UserRole,
}

impl Responder for UserJwtPayload {
    type Body = BoxBody;

    #[inline]
    fn respond_to(self, req: &actix_web::HttpRequest) -> actix_web::HttpResponse<Self::Body> {
        DataBody::new(self, "Success").respond_to(req)
    }
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub enum InvalidationReason {
    PasswordChanged,
    UserRequest,
    TooManyAuthFailures,
    UserDeleted,
    PermissionChanged,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct InvalidationData {
    pub date: u64,
    pub reason: InvalidationReason,
}

impl InvalidationData {
    fn new(date: u64, reason: InvalidationReason) -> Self {
        Self { date, reason }
    }
}

pub enum InvalidatedResult {
    Is(InvalidationData),
    Not,
}

impl UserJwtPayload {
    fn new(id: String, username: String, email: String, role: Option<UserRole>) -> Self {
        let now = now_unix_sec();

        Self {
            sub: id,
            username,
            email,
            exp: (now + (JWT_TOKEN_DURATION as u64)),
            iat: now,
            role: match role {
                Some(r) => r,
                None => UserRole::Common,
            },
        }
    }
}

const JWT_TOKEN_DURATION: usize = 3600;

#[async_trait]
pub trait AuthRepository: Sync + Send {
    async fn generate_token(
        &self,
        id: String,
        email: String,
        username: String,
        role: UserRole,
    ) -> Result<String, ApiError>;

    async fn add_invalidation(
        &self,
        id: String,
        reason: InvalidationReason,
    ) -> Result<(), ApiError>;

    async fn is_under_invalidation(&self, id: String) -> Result<InvalidatedResult, ApiError>;
    async fn decode_token(&self, token: String) -> Result<UserJwtPayload, ApiError>;
    async fn auth_user(&self, email: String, password: String) -> Result<String, ApiError>;
}

pub struct AuthService {
    db: &'static DatabaseConnection,
    cs: &'static CacheService,
    enc_key: EncodingKey,
    dec_key: DecodingKey,
    validation: Validation,
}

pub const JWT_ALGORITHM: Algorithm = Algorithm::EdDSA;

impl AuthService {
    pub fn new(
        db: &'static DatabaseConnection,
        cs: &'static CacheService,
        enc_key: EncodingKey,
        dec_key: DecodingKey,
    ) -> Self {
        let validation = Validation::new(JWT_ALGORITHM);

        Self {
            enc_key,
            dec_key,
            db,
            validation,
            cs,
        }
    }
}

#[async_trait]
impl AuthRepository for AuthService {
    async fn generate_token(
        &self,
        id: String,
        email: String,
        username: String,
        role: UserRole,
    ) -> Result<String, ApiError> {
        let claims = UserJwtPayload::new(id, username, email, Some(role));

        let key = self.enc_key.clone();

        spawn_blocking(move || jsonwebtoken::encode(&Header::new(JWT_ALGORITHM), &claims, &key))
            .await
            .or_else(|e| {
                log::error!(target: "tokio_runtime_error", "{}", e);
                Err(ApiError::InternalServerError)
            })?
            .or_else(|e| {
                log::error!(target: "jwt_error", "{}", e);
                Err(ApiError::InternalServerError)
            })
    }

    async fn add_invalidation(
        &self,
        id: String,
        reason: InvalidationReason,
    ) -> Result<(), ApiError> {
        let id = "invalidation/".to_string() + id.as_str();

        let payload = InvalidationData::new(now_unix_sec(), reason);
        let payload = serde_json::to_string(&payload).or_else(|e| {
            log::error!("Failed to encode invalidation data: {}", e);
            Err(ApiError::InternalServerError)
        })?;

        self.cs
            .set_ttl(id, payload, JWT_TOKEN_DURATION + 30)
            .await?;

        Ok(())
    }

    // Implement later
    async fn is_under_invalidation(&self, id: String) -> Result<InvalidatedResult, ApiError> {
        let id = "invalidation/".to_string() + id.as_str();

        let res = self.cs.get(id).await?;

        if let Some(invalidated) = res {
            let payload: InvalidationData =
                serde_json::from_str(invalidated.as_str()).or_else(|e| {
                    log::error!("Failed to parse cached invalidation data: {}", e);
                    Err(ApiError::InternalServerError)
                })?;

            Ok(InvalidatedResult::Is(payload))
        } else {
            Ok(InvalidatedResult::Not)
        }
    }

    async fn decode_token(&self, token: String) -> Result<UserJwtPayload, ApiError> {
        let key = self.dec_key.clone();
        let validation = self.validation.clone();

        let token = spawn_blocking(move || {
            let token: TokenData<UserJwtPayload> =
                jsonwebtoken::decode(token.as_str(), &key, &validation).or_else(|e| {
                    Err(match e.kind() {
                        JwtErrorKind::ExpiredSignature => ApiError::ExpiredAuthToken,
                        _ => ApiError::InvalidAuthToken,
                    })
                })?;

            if token.header.alg != JWT_ALGORITHM {
                Err(ApiError::InvalidAuthToken)
            } else if token.claims.exp < now_unix_sec() {
                Err(ApiError::ExpiredAuthToken)
            } else {
                Ok(token.claims)
            }
        })
        .await
        .or_else(|e| {
            log::error!(target: "tokio_runtime_error", "{}", e);
            Err(ApiError::InternalServerError)
        })??;

        match self.is_under_invalidation(token.sub.clone()).await? {
            InvalidatedResult::Is(data) => {
                if data.date + 10 < token.iat {
                    Ok(token)
                } else {
                    Err(ApiError::UserUnderTokenInvalidation(data.reason))
                }
            }
            InvalidatedResult::Not => Ok(token),
        }
    }

    async fn auth_user(&self, email: String, password: String) -> Result<String, ApiError> {
        let user = UserEntity::find()
            .filter(UserColumn::Email.eq(email.clone()))
            .column(UserColumn::Username)
            .column(UserColumn::Password)
            .one(self.db)
            .await
            .or_else(|_| Err(ApiError::UserUnauthorized))?;

        let user = match user {
            Some(user) => user,
            None => return Err(ApiError::UserUnauthorized),
        };

        let can_auth = spawn_blocking(move || {
            bcrypt::verify(password, user.password.as_str()).or_else(|e| {
                log::error!(target: "bcrypt_error", "{}", e);

                Err(ApiError::InternalServerError)
            })
        })
        .await
        .or_else(|e| {
            log::error!(target: "tokio_runtime_error", "{}", e);
            Err(ApiError::InternalServerError)
        })?;

        let can_auth = match can_auth {
            Ok(v) => v,
            Err(err) => return Err(err),
        };

        if can_auth {
            self.generate_token(user.id, email, user.username, user.role)
                .await
        } else {
            Err(ApiError::UserUnauthorized)
        }
    }
}
