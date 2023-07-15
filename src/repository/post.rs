use super::cache::CacheService;
use crate::{
    error::ApiError,
    model::post::{ActiveModel, Column as PostColumn, Entity as PostEntity, Model as PostModel},
    utils::db::{random_post_id, timestamp_now, USER_ID_SIZE},
};
use sea_orm::{
    ActiveModelTrait, ColumnTrait, DatabaseConnection, DbErr, EntityTrait, QueryFilter, QueryOrder,
    QuerySelect, Set,
};
use serde::{Deserialize, Serialize};

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct CreatePostData {
    slug: String,
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

        if self.slug.len() > 64 || self.slug.len() < 12 {
            return Some(ApiError::InvalidPostSlugSize);
        }

        None
    }
}

pub struct PostRepository {
    db: &'static DatabaseConnection,
    cm: &'static CacheService,
}

impl PostRepository {
    pub fn new(db: &'static DatabaseConnection, cm: &'static CacheService) -> Self {
        Self { db, cm }
    }

    pub async fn create(
        &self,
        user_id: String,
        data: CreatePostData,
    ) -> Result<PostModel, ApiError> {
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
            slug: Set(data.slug),
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

    pub async fn get_by_id(&self, id: String) -> Result<PostModel, ApiError> {
        let result = PostEntity::find_by_id(id)
            // .inner_join(UserEntity)
            .one(self.db)
            .await
            .or_else(|e| {
                Err(match e {
                    DbErr::RecordNotFound(_) => ApiError::PostNotFound,
                    e => {
                        log::error!(target: "database_user_errors", "{}", e);
                        ApiError::InternalServerError
                    }
                })
            })?;

        if let Some(post) = result {
            Ok(post)
        } else {
            Err(ApiError::PostNotFound)
        }
    }

    pub async fn get_user_posts(&self, id: String, limit: u64) -> Result<Vec<PostModel>, ApiError> {
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

        Ok(result)
    }

    pub async fn get_recomendation(&self, limit: usize) -> Result<Vec<PostModel>, ApiError> {
        let result = self.cm.get("post-recomendation".to_string()).await?;

        if let Some(cache) = result {
            let mut cache: Vec<PostModel> = serde_json::from_str(cache.as_str()).or_else(|e| {
                log::error!("Failed to decode cached recomendation: {}", e);
                Err(ApiError::InternalServerError)
            })?;

            if limit <= cache.len() {
                cache = cache[0..limit].to_vec()
            }

            Ok(cache)
        } else {
            let mut result = PostEntity::find()
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

            if limit <= result.len() {
                result = result[0..limit].to_vec()
            }

            Ok(result)
        }
    }
}
