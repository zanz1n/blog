use super::user::UserError;
use crate::model::user::{Column as UserColumn, Entity as UserEntity};
use jsonwebtoken::{Algorithm, EncodingKey, Header};
use sea_orm::DatabaseConnection;
use sea_orm::{ColumnTrait, EntityTrait, QueryFilter, QuerySelect};
use serde::{Deserialize, Serialize};
use tokio::task::spawn_blocking;

#[derive(Debug, Serialize, Deserialize)]
pub struct UserJwtPayload {
    #[serde(rename = "userId")]
    id: String,
    username: String,
    email: String,
}

impl UserJwtPayload {
    fn new(id: String, username: String, email: String) -> Self {
        Self {
            id,
            username,
            email,
        }
    }
}

pub struct AuthProvider {
    db: &'static DatabaseConnection,
    key: EncodingKey,
}

pub const JWT_ALGORITHM: Algorithm = Algorithm::HS512;

impl AuthProvider {
    pub fn new(db: &'static DatabaseConnection, key: EncodingKey) -> Self {
        Self { key, db }
    }

    pub async fn generate_token(
        &self,
        id: String,
        email: String,
        username: String,
    ) -> Result<String, UserError> {
        let claims = UserJwtPayload::new(id, username, email);

        let key = self.key.clone();

        spawn_blocking(move || jsonwebtoken::encode(&Header::new(JWT_ALGORITHM), &claims, &key))
            .await
            .or_else(|e| {
                log::error!(target: "tokio_runtime_error", "{}", e);
                Err(UserError::InternalServerError)
            })?
            .or_else(|_| Err(UserError::InternalServerError))
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
