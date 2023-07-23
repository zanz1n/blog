use super::cache::{CacheRepository, CacheService};
use crate::{
    error::ApiError,
    model::post::{
        ActiveModel, Column as PostColumn, Entity as PostEntity, Model as PostModel, PostWithUser,
    },
    model::user::Entity as UserEntity,
    utils::db::{random_post_id, sanitize_posts_job, timestamp_now, POST_ID_SIZE, USER_ID_SIZE},
};
use async_trait::async_trait;
use sea_orm::{
    ActiveModelTrait, ColumnTrait, DatabaseConnection, DbErr, EntityTrait, QueryFilter, QueryOrder,
    QuerySelect, Set,
};
use serde::{Deserialize, Serialize};

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct CreatePostData {
    title: String,
    content: String,
    #[serde(rename = "thumbImage")]
    thumb_image: Option<String>,
}

impl CreatePostData {
    pub fn is_valid(&self) -> Option<ApiError> {
        if let Some(thumb_image) = self.thumb_image.as_ref() {
            if thumb_image.len() > 128 {
                return Some(ApiError::InvalidPostThumbIdSize);
            }
        }

        if self.title.len() > 192 || self.title.len() < 12 {
            return Some(ApiError::InvalidPostTitleSize);
        }

        None
    }
}

#[async_trait]
pub trait PostRepository: Sync + Send {
    async fn create(&self, user_id: String, data: CreatePostData) -> Result<PostModel, ApiError>;
    async fn get_by_id(&self, id: String) -> Result<PostWithUser, ApiError>;
    async fn get_content(&self, id: String) -> Result<String, ApiError>;
    async fn get_user_posts(
        &self,
        id: String,
        limit: u64,
        offset: u64,
    ) -> Result<Vec<PostModel>, ApiError>;
    async fn get_recomendation(
        &self,
        limit: usize,
        cursor: usize,
    ) -> Result<Vec<PostModel>, ApiError>;
}

pub struct PostService {
    db: &'static DatabaseConnection,
    cm: &'static CacheService,
}

impl PostService {
    pub fn new(db: &'static DatabaseConnection, cm: &'static CacheService) -> Self {
        Self { db, cm }
    }
}

#[async_trait]
impl PostRepository for PostService {
    async fn create(&self, user_id: String, data: CreatePostData) -> Result<PostModel, ApiError> {
        if let Some(err) = data.is_valid() {
            return Err(err);
        }

        let now = match timestamp_now() {
            Some(v) => v,
            None => {
                log::warn!("Failed to get timestamp");
                return Err(ApiError::InternalServerError);
            }
        };

        let post = ActiveModel {
            id: Set(random_post_id()),
            title: Set(data.title),
            thumb_image: Set(data.thumb_image),
            content: Set(data.content),
            created_at: Set(now),
            updated_at: Set(now),
            user_id: Set(user_id),
        };

        let post = post
            .insert(self.db)
            .await
            .or_else(|_| Err(ApiError::InternalServerError))?;

        Ok(post)
    }

    async fn get_by_id(&self, id: String) -> Result<PostWithUser, ApiError> {
        if id.len() != POST_ID_SIZE {
            return Err(ApiError::InvalidPostIdSize);
        }

        let post_result = PostEntity::find_by_id(id).one(self.db).await.or_else(|e| {
            Err(match e {
                DbErr::RecordNotFound(_) => ApiError::PostNotFound,
                e => {
                    log::error!(target: "database_post_errors", "{}", e);
                    ApiError::InternalServerError
                }
            })
        })?;

        if let Some(post) = post_result {
            let user_result = UserEntity::find_by_id(post.user_id.clone())
                .one(self.db)
                .await
                .or_else(|e| {
                    log::error!(target: "database_user_errors", "{}", e);
                    Err(ApiError::InternalServerError)
                })?;

            Ok(PostWithUser::new(
                post,
                match user_result {
                    Some(v) => Some(v.to_sendable()),
                    None => None,
                },
            ))
        } else {
            Err(ApiError::PostNotFound)
        }
    }

    async fn get_user_posts(
        &self,
        id: String,
        limit: u64,
        offset: u64,
    ) -> Result<Vec<PostModel>, ApiError> {
        if id.len() != USER_ID_SIZE {
            return Err(ApiError::InvalidUserIdSize);
        }

        let mut limit = limit;

        if limit > 256 {
            limit = 256
        }

        let result = PostEntity::find()
            .filter(PostColumn::UserId.eq(id))
            .order_by_desc(PostColumn::CreatedAt)
            .offset(offset)
            .limit(limit)
            .all(self.db)
            .await
            .or_else(|err| match err {
                DbErr::RecordNotFound(_) => Ok(vec![]),
                e => {
                    log::error!(target: "database_user_errors", "{}", e);
                    Err(ApiError::InternalServerError)
                }
            })?;

        let result = sanitize_posts_job(result).await?;

        Ok(result)
    }

    async fn get_recomendation(
        &self,
        limit: usize,
        cursor: usize,
    ) -> Result<Vec<PostModel>, ApiError> {
        let result = self.cm.get("post-recomendation".to_string()).await?;

        let mut result = if let Some(cache) = result {
            serde_json::from_str::<Vec<PostModel>>(cache.as_str()).or_else(|e| {
                log::error!("Failed to decode cached recomendation: {}", e);
                Err(ApiError::InternalServerError)
            })?
        } else {
            let result = PostEntity::find()
                .order_by_desc(PostColumn::CreatedAt)
                .limit(256)
                .all(self.db)
                .await
                .or_else(|err| match err {
                    DbErr::RecordNotFound(_) => Ok(vec![]),
                    e => {
                        log::error!(target: "database_user_errors", "{}", e);
                        Err(ApiError::InternalServerError)
                    }
                })?;

            let result = sanitize_posts_job(result).await?;

            match serde_json::to_string(&result) {
                Ok(v) => {
                    _ = self
                        .cm
                        .set_ttl("post-recomendation".to_string(), v, 60)
                        .await
                        .or_else(|e| {
                            log::error!("Failed to cache post recomendations: {}", e);
                            Err(())
                        });
                }
                Err(e) => {
                    log::error!("Failed to encode cached recomendation: {}", e);
                }
            }

            result
        };

        if cursor >= result.len() {
            return Err(ApiError::PostNotFound);
        }

        if limit <= result.len() {
            result = result[cursor..limit].to_vec()
        }

        Ok(result)
    }

    async fn get_content(&self, id: String) -> Result<String, ApiError> {
        let result = PostEntity::find_by_id(id)
            .column(PostColumn::Content)
            .one(self.db)
            .await
            .or_else(|e| {
                Err(match e {
                    DbErr::RecordNotFound(_) => ApiError::PostNotFound,
                    _ => ApiError::InternalServerError,
                })
            })?;

        match result {
            Some(v) => Ok(v.content),
            None => Err(ApiError::PostNotFound),
        }
    }
}
