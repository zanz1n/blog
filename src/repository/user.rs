use crate::{
    error::ApiError,
    model::user::Entity as UserEntity,
    model::user::{ActiveModel, Model as UserModel, UserRole},
    utils::db::{db_to_user_error, hash_password, random_user_id, timestamp_now},
};
use async_trait::async_trait;
use sea_orm::{
    ActiveModelTrait, ActiveValue::NotSet, DatabaseConnection, EntityTrait, Set, Unchanged,
};
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateUserData {
    username: String,
    email: String,
    password: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateEmailData {
    username: String,
}

impl CreateUserData {
    fn is_valid(&self) -> Option<ApiError> {
        if self.password.len() < 6
            || self.password.contains(self.email.as_str())
            || self.password.contains(self.username.as_str())
            || self.password == "1234567"
        {
            Some(ApiError::WeakUserPasswordError)
        } else if self.username.len() > 42 {
            Some(ApiError::UsernameTooBig)
        } else if self.email.len() > 64 {
            Some(ApiError::UserEmailTooBig)
        } else {
            None
        }
    }
}

#[async_trait]
pub trait UserRepository {
    async fn get_by_id(&self, id: String) -> Result<UserModel, ApiError>;
    async fn create(&self, data: CreateUserData) -> Result<UserModel, ApiError>;
    async fn update_username(
        &self,
        id: String,
        data: UpdateEmailData,
    ) -> Result<UserModel, ApiError>;

    async fn delete(&self, id: String) -> Result<(), ApiError>;
}

pub struct UserService {
    db: &'static DatabaseConnection,
}

impl UserService {
    pub fn new(db: &'static DatabaseConnection) -> Self {
        Self { db }
    }
}

#[async_trait]
impl UserRepository for UserService {
    async fn get_by_id(&self, id: String) -> Result<UserModel, ApiError> {
        if id.len() > 18 {
            return Err(ApiError::InvalidUserIdSize);
        };

        let select_result = UserEntity::find_by_id(id).one(self.db).await;

        match select_result {
            Ok(result) => match result {
                Some(user) => Ok(user),
                None => Err(ApiError::UserNotFound),
            },
            Err(_) => Err(ApiError::InternalServerError),
        }
    }

    async fn create(&self, data: CreateUserData) -> Result<UserModel, ApiError> {
        match data.is_valid() {
            Some(err) => return Err(err),
            None => {}
        };

        let now = match timestamp_now() {
            Some(v) => v,
            None => {
                log::warn!("Failed to get timestamp");
                return Err(ApiError::InternalServerError);
            }
        };

        let password = hash_password(data.password).await?;

        let user = ActiveModel {
            id: Set(random_user_id()),
            email: Set(data.email),
            password: Set(password),
            username: Set(data.username),
            created_at: Set(now),
            updated_at: Set(now),
            role: Set(UserRole::Common),
        };

        let user = match user.insert(self.db).await {
            Ok(v) => v,
            Err(err) => return Err(db_to_user_error(err, ApiError::UserAlreadyExists)),
        };

        Ok(user)
    }

    async fn update_username(
        &self,
        id: String,
        data: UpdateEmailData,
    ) -> Result<UserModel, ApiError> {
        if id.len() > 18 {
            return Err(ApiError::InvalidUserIdSize);
        };

        let user = ActiveModel {
            id: Unchanged(id),
            created_at: NotSet,
            email: NotSet,
            password: NotSet,
            updated_at: NotSet,
            username: Set(data.username),
            role: NotSet,
        };

        log::info!("{:?}", user);

        let user = match user.update(self.db).await {
            Ok(u) => u,
            Err(err) => return Err(db_to_user_error(err, ApiError::UserNotFound)),
        };

        Ok(user)
    }

    async fn delete(&self, id: String) -> Result<(), ApiError> {
        if id.len() > 18 {
            return Err(ApiError::InvalidUserIdSize);
        };

        let user = ActiveModel {
            id: Unchanged(id.clone()),
            created_at: NotSet,
            email: NotSet,
            password: NotSet,
            updated_at: NotSet,
            username: NotSet,
            role: NotSet,
        };

        let result = match user.delete(self.db).await {
            Ok(r) => r,
            Err(_) => return Err(ApiError::InternalServerError),
        };

        match result.rows_affected {
            0 => Err(ApiError::UserNotFound),
            1 => Ok(()),
            i => {
                log::info!(
                    target: "database_user",
                    "User {} deletion affected other {} rows",
                    id,
                    i - 1
                );
                Ok(())
            }
        }
    }
}
