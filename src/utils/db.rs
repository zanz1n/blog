use std::time::{SystemTime, UNIX_EPOCH};

use sea_orm::{prelude::DateTime, DbErr};

use crate::repository::user::UserError;

const USER_ID_SIZE: usize = 18;
const POST_ID_SIZE: usize = 24;

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
pub fn db_to_user_error(db_err: DbErr, expect: UserError) -> UserError {
    log::info!(target: "database_user_errors", "{}", db_err.to_string());

    match db_err {
        DbErr::Exec(_) => expect,
        DbErr::Type(_) => UserError::InvalidData,
        DbErr::Query(_) => expect,
        DbErr::RecordNotFound(_) => UserError::NotFound,
        _ => UserError::InternalServerError,
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