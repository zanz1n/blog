use actix_web::{body::BoxBody, http::StatusCode, ResponseError};

use crate::utils::http::{serialize_response, ErrorResponseBody};

#[derive(Debug, thiserror::Error)]
pub enum ApiError {
    #[error("User could not be found")]
    UserNotFound,
    #[error("Users id's have a fixed size of 18 characters")]
    InvalidUserIdSize,
    #[error("Something went wrong while processing your request, try again later")]
    InternalServerError,
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
    #[error("Your jwt token does not contain valid metadata. Invalid")]
    InvalidAuthToken,
    #[error("Your jwt token is no longer valid. Expired")]
    ExpiredAuthToken,
    #[error("This route requires authorization but no headers or cookies was provided")]
    AuthorizationRequired,
    #[error("This route does not support sinature based authentication")]
    SignatureAuthNotSupported,
    #[error("The provided authorization header is not valid, ex: `Bearer <token>` or `Signature <token>`")]
    InvalidAuthHeaderFormat,
}

impl ResponseError for ApiError {
    fn status_code(&self) -> StatusCode {
        match self {
            Self::UserNotFound => StatusCode::NOT_FOUND,
            Self::InternalServerError => StatusCode::INTERNAL_SERVER_ERROR,
            Self::UserAlreadyExists => StatusCode::CONFLICT,
            Self::UserUnauthorized => StatusCode::UNAUTHORIZED,
            Self::AuthorizationRequired => StatusCode::UNAUTHORIZED,
            _ => StatusCode::BAD_REQUEST,
        }
    }

    fn error_response(&self) -> actix_web::HttpResponse<BoxBody> {
        serialize_response(&ErrorResponseBody::from(self), self.status_code())
    }
}