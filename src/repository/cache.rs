use std::io::{self, Error};

use crate::error::ApiError;
use deadpool_redis::{
    redis::{AsyncCommands, Expiry},
    Config, Connection, Pool, Runtime,
};

pub struct CacheService {
    client: Pool,
}

impl CacheService {
    pub fn new(uri: String) -> Result<Self, Error> {
        let client = Config::from_url(uri)
            .create_pool(Some(Runtime::Tokio1))
            .or_else(|e| Err(Error::new(io::ErrorKind::ConnectionRefused, e)))?;

        Ok(Self { client })
    }

    pub async fn get_conn(&self) -> Result<Connection, ApiError> {
        self.client
            .get()
            .await
            .or_else(|_| Err(ApiError::InternalServerError))
    }

    pub async fn get(&self, key: String) -> Result<Option<String>, ApiError> {
        let mut conn = self.get_conn().await?;

        let value: String = conn
            .get(key)
            .await
            .or_else(|_| Err(ApiError::InternalServerError))?;

        Ok(Some(value))
    }

    pub async fn get_ttl(&self, key: String, ttl: usize) -> Result<Option<String>, ApiError> {
        let mut conn = self.get_conn().await?;

        let value: String = conn
            .get_ex(key, Expiry::EX(ttl))
            .await
            .or_else(|_| Err(ApiError::InternalServerError))?;

        Ok(Some(value))
    }

    pub async fn set(&self, key: String, value: String) -> Result<(), ApiError> {
        let mut conn = self.get_conn().await?;

        conn.set(key, value)
            .await
            .or_else(|_| Err(ApiError::InternalServerError))?;

        Ok(())
    }

    pub async fn set_ttl(&self, key: String, value: String, ttl: usize) -> Result<(), ApiError> {
        let mut conn = self.get_conn().await?;

        conn.set_ex(key, value, ttl)
            .await
            .or_else(|_| Err(ApiError::InternalServerError))?;

        Ok(())
    }
}
