use super::user::UserError;
use crate::model::user::{Column as UserColumn, Entity as UserEntity};
use crate::utils::generic::now_unix_sec;
use jsonwebtoken::{Algorithm, DecodingKey, EncodingKey, Header, TokenData, Validation};
use sea_orm::DatabaseConnection;
use sea_orm::{ColumnTrait, EntityTrait, QueryFilter, QuerySelect};
use serde::{Deserialize, Serialize};
use tokio::task::spawn_blocking;

#[derive(Debug, Serialize, Deserialize)]
pub struct UserJwtPayload {
    sub: String,
    username: String,
    email: String,
    exp: u64,
    iat: u64,
}

impl UserJwtPayload {
    fn new(id: String, username: String, email: String) -> Self {
        let now = now_unix_sec();

        Self {
            sub: id,
            username,
            email,
            exp: (now + (60 * 60)),
            iat: now,
        }
    }
}

pub struct AuthProvider {
    db: &'static DatabaseConnection,
    enc_key: EncodingKey,
    dec_key: DecodingKey,
    validation: Validation,
}

pub const JWT_ALGORITHM: Algorithm = Algorithm::EdDSA;

impl AuthProvider {
    pub fn new(
        db: &'static DatabaseConnection,
        enc_key: EncodingKey,
        dec_key: DecodingKey,
    ) -> Self {
        let validation = Validation::new(JWT_ALGORITHM);

        Self {
            enc_key,
            dec_key,
            db,
            validation,
        }
    }

    pub async fn generate_token(
        &self,
        id: String,
        email: String,
        username: String,
    ) -> Result<String, UserError> {
        let claims = UserJwtPayload::new(id, username, email);

        let key = self.enc_key.clone();

        spawn_blocking(move || jsonwebtoken::encode(&Header::new(JWT_ALGORITHM), &claims, &key))
            .await
            .or_else(|e| {
                log::error!(target: "tokio_runtime_error", "{}", e);
                Err(UserError::InternalServerError)
            })?
            .or_else(|e| {
                log::error!(target: "jwt_error", "{}", e);
                Err(UserError::InternalServerError)
            })
    }

    // Implement later
    pub async fn is_under_invalidation(&self, _id: String) -> Result<bool, UserError> {
        Ok(false)
    }

    pub async fn decode_token(&self, token: String) -> Result<UserJwtPayload, UserError> {
        let key = self.dec_key.clone();
        let validation = self.validation.clone();

        let token = spawn_blocking(move || {
            let token: TokenData<UserJwtPayload> =
                jsonwebtoken::decode(token.as_str(), &key, &validation).or_else(|e| {
                    log::error!(target: "jwt_error", "{}", e);
                    Err(UserError::InternalServerError)
                })?;

            if token.header.alg != JWT_ALGORITHM {
                Err(UserError::InvalidAuthToken)
            } else if token.claims.exp < now_unix_sec() {
                Err(UserError::ExpiredAuthToken)
            } else {
                Ok(token.claims)
            }
        })
        .await
        .or_else(|e| {
            log::error!(target: "tokio_runtime_error", "{}", e);
            Err(UserError::InternalServerError)
        })??;

        if self.is_under_invalidation(token.sub.clone()).await? {
            Err(UserError::InvalidAuthToken)
        } else {
            Ok(token)
        }
    }

    pub async fn auth_user(&self, email: String, password: String) -> Result<String, UserError> {
        let user = UserEntity::find()
            .filter(UserColumn::Email.eq(email.clone()))
            .column(UserColumn::Username)
            .column(UserColumn::Password)
            .one(self.db)
            .await
            .or_else(|_| Err(UserError::Unauthorized))?;

        let user = match user {
            Some(user) => user,
            None => return Err(UserError::Unauthorized),
        };

        let can_auth = spawn_blocking(move || {
            bcrypt::verify(password, user.password.as_str()).or_else(|e| {
                log::error!(target: "bcrypt_error", "{}", e);

                Err(UserError::InternalServerError)
            })
        })
        .await
        .or_else(|e| {
            log::error!(target: "tokio_runtime_error", "{}", e);
            Err(UserError::InternalServerError)
        })?;

        let can_auth = match can_auth {
            Ok(v) => v,
            Err(err) => return Err(err),
        };

        if can_auth {
            self.generate_token(user.id, email, user.username).await
        } else {
            Err(UserError::Unauthorized)
        }
    }
}
