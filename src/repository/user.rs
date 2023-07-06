use crate::{
    model::user::Entity as UserEntity,
    model::user::{ActiveModel, Model as UserModel},
    utils::{
        db::{random_user_id, timestamp_now},
        http::{serialize_response, ErrorResponseBody},
    },
};
use actix_web::{body::BoxBody, http::StatusCode, ResponseError};
use sea_orm::{ActiveModelTrait, DatabaseConnection, DbErr, EntityTrait, Set};
use serde::{Deserialize, Serialize};

#[derive(Debug, thiserror::Error)]
pub enum UserError {
    #[error("User could not be found")]
    NotFound,
    #[error("Users id's have a fixed size of 18 characters")]
    InvalidIdSize,
    #[error("Something went wrong while processing your request, try again later")]
    InternalServerError,
    #[error(
        "Your password length must be greater than 6 and must not contain your username or email"
    )]
    WeakPasswordError,
    #[error("Usernames must be less than 42 characters")]
    UsernameTooBig,
    #[error("Emails must be less than 64 characters")]
    EmailTooBig,
    #[error("User payload contain invalid fields")]
    InvalidData,
    #[error("User already exists, maybe try a different email")]
    AlreadyExists,
}

impl ResponseError for UserError {
    fn status_code(&self) -> StatusCode {
        match self {
            Self::NotFound => StatusCode::NOT_FOUND,
            Self::InternalServerError => StatusCode::INTERNAL_SERVER_ERROR,
            Self::AlreadyExists => StatusCode::CONFLICT,
            _ => StatusCode::BAD_REQUEST,
        }
    }

    fn error_response(&self) -> actix_web::HttpResponse<BoxBody> {
        serialize_response(&ErrorResponseBody::from(self), self.status_code())
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateUserData {
    username: String,
    email: String,
    password: String,
}

impl CreateUserData {
    fn is_valid(&self) -> Option<UserError> {
        if self.password.len() < 6
            || self.password.contains(self.email.as_str())
            || self.password.contains(self.username.as_str())
            || self.password == "1234567"
        {
            Some(UserError::WeakPasswordError)
        } else if self.username.len() > 42 {
            Some(UserError::UsernameTooBig)
        } else if self.email.len() > 64 {
            Some(UserError::EmailTooBig)
        } else {
            None
        }
    }
}

pub struct UserRepository {
    db: &'static DatabaseConnection,
}

impl UserRepository {
    pub fn new(db: &'static DatabaseConnection) -> Self {
        Self { db }
    }

    pub async fn get_by_id(&self, id: String) -> Result<UserModel, UserError> {
        if id.len() > 18 {
            return Err(UserError::InvalidIdSize);
        };

        let select_result = UserEntity::find_by_id(id).one(self.db).await;

        match select_result {
            Ok(result) => match result {
                Some(user) => Ok(user),
                None => Err(UserError::NotFound),
            },
            Err(e) => match e {
                DbErr::RecordNotFound(_) => Err(UserError::NotFound),
                _ => Err(UserError::InternalServerError),
            },
        }
    }

    pub async fn create(&self, data: CreateUserData) -> Result<UserModel, UserError> {
        match data.is_valid() {
            Some(err) => return Err(err),
            None => {}
        };

        let now = match timestamp_now() {
            Some(v) => v,
            None => {
                log::warn!("Failed to get timestamp");
                return Err(UserError::InternalServerError);
            }
        };

        let user = ActiveModel {
            id: Set(random_user_id()),
            email: Set(data.email),
            password: Set(data.password),
            username: Set(data.username),
            created_at: Set(now),
            updated_at: Set(now),
        };

        let user = match user.insert(self.db).await {
            Ok(v) => v,
            Err(err) => {
                return match err {
                    DbErr::Exec(_) => Err(UserError::InvalidData),
                    DbErr::Type(_) => Err(UserError::InvalidData),
                    DbErr::Query(_) => Err(UserError::AlreadyExists),
                    _ => Err(UserError::InternalServerError),
                };
            }
        };

        Ok(user)
    }
}
