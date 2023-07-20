use crate::{repository::auth::InvalidationReason, utils::http::serialize_response};
use actix_web::{body::BoxBody, http::StatusCode, ResponseError};
use serde::{Deserialize, Serialize};
use std::{fmt::Display, io};

#[derive(Debug, thiserror::Error)]
pub enum ApiError {
    #[error("User could not be found")]
    UserNotFound,
    #[error("Users id's have a fixed size of 18 characters")]
    InvalidUserIdSize,
    #[error("Something went wrong while processing your request, try again later")]
    InternalServerError,
    #[error("Failed to decode request body, invalid payload!")]
    InvalidBodyPayload,
    #[error(
        "Your password length must be greater than 6 and must not contain your username or email"
    )]
    WeakUserPasswordError,
    #[error("Usernames must be less than 42 characters")]
    UsernameTooBig,
    #[error("Emails must be less than 64 characters")]
    UserEmailTooBig,
    #[error("User payload contain invalid fields")]
    InvalidUserData,
    #[error("User already exists, maybe try a different email")]
    UserAlreadyExists,
    #[error("Password do not match or user doesn't exist")]
    UserUnauthorized,
    #[error("Your jwt token does not contain valid metadata")]
    InvalidAuthToken,
    #[error("Your jwt token is no longer valid, expired!")]
    ExpiredAuthToken,
    #[error("This route requires authorization but no headers or cookies was provided")]
    AuthorizationRequired,
    #[error("This route does not support sinature based authentication")]
    SignatureAuthNotSupported,
    #[error("The provided authorization header is not valid, ex: `Bearer <token>` or `Signature <token>`")]
    InvalidAuthHeaderFormat,
    #[error("Your authentication token is not longer valid, please login again")]
    UserUnderTokenInvalidation(InvalidationReason),
    #[error("You can only mutate information if you own them or if you are a mod/admin")]
    DataMutationDenied,
    #[error("Post could not be found")]
    PostNotFound,
    #[error("Users id's have a fixed size of 24 characters")]
    InvalidPostIdSize,
    #[error("Posts 'thumbImage' prop must be a nullable string up to 128 characters")]
    InvalidPostThumbIdSize,
    #[error("Posts 'slug' prop must at least 12 and up to 128 characters")]
    InvalidPostSlugSize,
}

impl From<&ApiError> for usize {
    fn from(value: &ApiError) -> usize {
        match value {
            ApiError::UserNotFound => 4041,
            ApiError::InvalidUserIdSize => 4001,
            ApiError::InternalServerError => 5000,
            ApiError::InvalidBodyPayload => 4000,
            ApiError::WeakUserPasswordError => 4002,
            ApiError::UsernameTooBig => 4003,
            ApiError::UserEmailTooBig => 4004,
            ApiError::InvalidUserData => 4005,
            ApiError::UserAlreadyExists => 4070,
            ApiError::UserUnauthorized => 4010,
            ApiError::InvalidAuthToken => 4011,
            ApiError::ExpiredAuthToken => 4012,
            ApiError::AuthorizationRequired => 4013,
            ApiError::SignatureAuthNotSupported => 4006,
            ApiError::InvalidAuthHeaderFormat => 4014,
            ApiError::DataMutationDenied => 4016,
            ApiError::PostNotFound => 4042,
            ApiError::InvalidPostIdSize => 4007,
            ApiError::InvalidPostThumbIdSize => 4008,
            ApiError::InvalidPostSlugSize => 4009,
            ApiError::UserUnderTokenInvalidation(r) => match r {
                InvalidationReason::PasswordChanged => 40151,
                InvalidationReason::PermissionChanged => 40152,
                InvalidationReason::TooManyAuthFailures => 40153,
                InvalidationReason::UserDeleted => 40154,
                InvalidationReason::UserRequest => 40155,
            },
        }
    }
}

impl ApiError {
    pub fn from_code(value: usize) -> Result<Self, std::io::Error> {
        match value {
            4041 => Ok(Self::UserNotFound),
            4001 => Ok(Self::InvalidUserIdSize),
            5000 => Ok(Self::InternalServerError),
            4000 => Ok(Self::InvalidBodyPayload),
            4002 => Ok(Self::WeakUserPasswordError),
            4003 => Ok(Self::UsernameTooBig),
            4004 => Ok(Self::UserEmailTooBig),
            4005 => Ok(Self::InvalidUserData),
            4070 => Ok(Self::UserAlreadyExists),
            4010 => Ok(Self::UserUnauthorized),
            4011 => Ok(Self::InvalidAuthToken),
            4012 => Ok(Self::ExpiredAuthToken),
            4013 => Ok(Self::AuthorizationRequired),
            4006 => Ok(Self::SignatureAuthNotSupported),
            4014 => Ok(Self::InvalidAuthHeaderFormat),
            4016 => Ok(Self::DataMutationDenied),
            4042 => Ok(Self::PostNotFound),
            4007 => Ok(Self::InvalidPostIdSize),
            4008 => Ok(Self::InvalidPostThumbIdSize),
            4009 => Ok(Self::InvalidPostSlugSize),
            40151 => Ok(Self::UserUnderTokenInvalidation(
                InvalidationReason::PasswordChanged,
            )),
            40152 => Ok(Self::UserUnderTokenInvalidation(
                InvalidationReason::PermissionChanged,
            )),
            40153 => Ok(Self::UserUnderTokenInvalidation(
                InvalidationReason::TooManyAuthFailures,
            )),
            40154 => Ok(Self::UserUnderTokenInvalidation(
                InvalidationReason::UserDeleted,
            )),
            40155 => Ok(Self::UserUnderTokenInvalidation(
                InvalidationReason::UserRequest,
            )),
            _ => Err(io::Error::new(
                io::ErrorKind::InvalidData,
                "Invalid error code, no error match this code",
            )),
        }
    }
}

impl ResponseError for ApiError {
    fn status_code(&self) -> StatusCode {
        match self {
            Self::UserNotFound => StatusCode::NOT_FOUND,
            Self::PostNotFound => StatusCode::NOT_FOUND,
            Self::InternalServerError => StatusCode::INTERNAL_SERVER_ERROR,
            Self::UserAlreadyExists => StatusCode::CONFLICT,
            Self::UserUnauthorized => StatusCode::UNAUTHORIZED,
            Self::AuthorizationRequired => StatusCode::UNAUTHORIZED,
            Self::UserUnderTokenInvalidation(_) => StatusCode::UNAUTHORIZED,
            Self::DataMutationDenied => StatusCode::UNAUTHORIZED,
            Self::InvalidAuthToken => StatusCode::UNAUTHORIZED,
            Self::InvalidAuthHeaderFormat => StatusCode::UNAUTHORIZED,
            Self::ExpiredAuthToken => StatusCode::UNAUTHORIZED,
            _ => StatusCode::BAD_REQUEST,
        }
    }

    fn error_response(&self) -> actix_web::HttpResponse<BoxBody> {
        serialize_response(
            &ErrorResponseBody::new(self, usize::from(self)),
            self.status_code(),
        )
    }
}

#[derive(Default, Debug, Serialize, Deserialize)]
pub struct ErrorResponseBody {
    error: String,
    #[serde(rename = "errorCode")]
    error_code: usize,
}

impl ErrorResponseBody {
    pub fn new<T: ToString>(el: T, code: usize) -> Self {
        Self {
            error: el.to_string(),
            error_code: code,
        }
    }
}

impl Display for ErrorResponseBody {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(format!("{:?}", self).as_str())
    }
}

impl ResponseError for ErrorResponseBody {
    fn status_code(&self) -> StatusCode {
        StatusCode::BAD_REQUEST
    }

    fn error_response(&self) -> actix_web::HttpResponse<BoxBody> {
        serialize_response(&self, self.status_code())
    }
}
