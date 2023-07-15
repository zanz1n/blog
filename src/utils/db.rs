use crate::error::ApiError;
use sea_orm::{prelude::DateTime, DbErr};
use std::time::{SystemTime, UNIX_EPOCH};
use tokio::task::spawn_blocking;

pub const USER_ID_SIZE: usize = 18;
pub const POST_ID_SIZE: usize = 24;

#[inline]
pub fn random_id(size: usize) -> String {
    nanoid::nanoid!(size)
}

#[inline]
pub fn random_user_id() -> String {
    nanoid::nanoid!(USER_ID_SIZE)
}

#[inline]
pub fn random_post_id() -> String {
    nanoid::nanoid!(POST_ID_SIZE)
}

#[inline]
pub fn db_to_user_error(db_err: DbErr, expect: ApiError) -> ApiError {
    log::info!(target: "database_user_errors", "{}", db_err.to_string());

    match db_err {
        DbErr::Exec(_) => expect,
        DbErr::Type(_) => ApiError::InvalidUserData,
        DbErr::Query(_) => expect,
        DbErr::RecordNotFound(_) => ApiError::UserNotFound,
        _ => ApiError::InternalServerError,
    }
}

#[inline]
pub fn now_unix_i64() -> Option<i64> {
    match SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_millis()
        .try_into()
    {
        Ok(v) => Some(v),
        Err(_) => None,
    }
}

#[inline]
pub fn timestamp_now() -> Option<DateTime> {
    let now_ms = now_unix_i64()?;
    DateTime::from_timestamp_millis(now_ms)
}

pub async fn hash_password(password: String) -> Result<String, ApiError> {
    spawn_blocking(|| bcrypt::hash(password, 12))
        .await
        .or_else(|e| {
            log::error!(target: "tokio_runtime_error", "{}", e);
            Err(ApiError::InternalServerError)
        })?
        .or_else(|e| {
            log::error!(target: "bcrypt_error", "{}", e);
            Err(ApiError::InternalServerError)
        })
}
