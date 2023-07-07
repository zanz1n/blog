use super::user::{UserError, UserRepository};
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
    user_repo: UserRepository,
}

pub const JWT_ALGORITHM: Algorithm = Algorithm::HS512;

impl AuthProvider {
    pub fn new(db: &'static DatabaseConnection, key: EncodingKey) -> Self {
        Self {
            key,
            db,
            user_repo: UserRepository::new(db),
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

        let can_auth =
            spawn_blocking(move || bcrypt::verify(password, user.password.as_str())).await;

        let can_auth = match can_auth {
            Ok(result) => match result {
                Ok(v) => v,
                Err(_) => return Err(UserError::InternalServerError),
            },
            Err(_) => return Err(UserError::InternalServerError),
        };

        if can_auth {
            let claims = UserJwtPayload::new(user.id, user.username, email);

            let key = self.key.clone();

            let token_result = spawn_blocking(move || {
                jsonwebtoken::encode(&Header::new(JWT_ALGORITHM), &claims, &key)
            })
            .await;

            match token_result {
                Ok(result) => result.or_else(|_| Err(UserError::InternalServerError)),
                Err(_) => Err(UserError::InternalServerError),
            }
        } else {
            Err(UserError::Unauthorized)
        }
    }
}
