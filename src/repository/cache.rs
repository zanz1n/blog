use crate::error::ApiError;
use async_trait::async_trait;
use deadpool_redis::{
    redis::{self, AsyncCommands, Expiry},
    Config, Connection, Pool, Runtime,
};
use std::io::{self, Error};

#[async_trait]
pub trait CacheRepository: Sync + Send {
    async fn get(&self, key: String) -> Result<Option<String>, ApiError>;
    async fn get_ttl(&self, key: String, ttl: usize) -> Result<Option<String>, ApiError>;
    async fn set(&self, key: String, value: String) -> Result<(), ApiError>;
    async fn set_ttl(&self, key: String, value: String, ttl: usize) -> Result<(), ApiError>;
}

#[derive(Clone)]
pub struct CacheService {
    client: Pool,
}

impl CacheService {
    pub async fn new(uri: String) -> Result<Self, Error> {
        let client = Config::from_url(uri)
            .create_pool(Some(Runtime::Tokio1))
            .or_else(|e| Err(Error::new(io::ErrorKind::ConnectionRefused, e)))?;

        let mut conn = client
            .get()
            .await
            .or_else(|e| Err(Error::new(io::ErrorKind::ConnectionRefused, e)))?;

        redis::cmd("PING")
            .query_async::<_, ()>(&mut conn)
            .await
            .or_else(|e| Err(Error::new(io::ErrorKind::ConnectionRefused, e)))?;

        Ok(Self { client })
    }

    pub async fn get_conn(&self) -> Result<Connection, ApiError> {
        self.client.get().await.or_else(|e| {
            log::error!("Failed to get_conn(): {}", e);
            Err(ApiError::InternalServerError)
        })
    }
}

#[async_trait]
impl CacheRepository for CacheService {
    async fn get(&self, key: String) -> Result<Option<String>, ApiError> {
        let mut conn = self.get_conn().await?;

        let value: Option<String> = conn.get(key).await.or_else(|e| {
            log::error!("Failed to get(): {}", e);
            Err(ApiError::InternalServerError)
        })?;

        Ok(value)
    }

    async fn get_ttl(&self, key: String, ttl: usize) -> Result<Option<String>, ApiError> {
        let mut conn = self.get_conn().await?;

        let value: Option<String> = conn.get_ex(key, Expiry::EX(ttl)).await.or_else(|e| {
            log::error!("Failed to get_ttl(): {}", e);
            Err(ApiError::InternalServerError)
        })?;

        Ok(value)
    }

    async fn set(&self, key: String, value: String) -> Result<(), ApiError> {
        let mut conn = self.get_conn().await?;

        conn.set(key, value).await.or_else(|e| {
            log::error!("Failed to set(): {}", e);
            Err(ApiError::InternalServerError)
        })?;

        Ok(())
    }

    async fn set_ttl(&self, key: String, value: String, ttl: usize) -> Result<(), ApiError> {
        let mut conn = self.get_conn().await?;

        conn.set_ex(key, value, ttl).await.or_else(|e| {
            log::error!("Failed to set_ttl(): {}", e);
            Err(ApiError::InternalServerError)
        })?;

        Ok(())
    }
}
