use sea_orm::{DatabaseConnection, DbErr, EntityTrait, QueryOrder, QuerySelect};

use crate::{
    error::ApiError,
    model::{
        post::{Column as PostColumn, Entity as PostEntity, Model as PostModel},
        user::Entity as UserEntity,
    },
};

use super::cache::CacheService;

pub struct PostRepository {
    db: &'static DatabaseConnection,
    cm: &'static CacheService,
}

impl PostRepository {
    pub async fn get_by_id(&self, id: String) -> Result<PostModel, ApiError> {
        let result = PostEntity::find_by_id(id)
            .inner_join(UserEntity)
            .one(self.db)
            .await
            .or_else(|e| {
                Err(match e {
                    DbErr::RecordNotFound(_) => ApiError::PostNotFound,
                    _ => ApiError::InternalServerError,
                })
            })?;

        if let Some(post) = result {
            Ok(post)
        } else {
            Err(ApiError::PostNotFound)
        }
    }

    pub async fn get_recomendation(&self) -> Result<Vec<PostModel>, ApiError> {
        let result = self.cm.get("".to_string()).await?;

        if let Some(cache) = result {
            let cache: Vec<PostModel> = serde_json::from_str(cache.as_str()).or_else(|e| {
                log::error!("Failed to decode cached recomendation: {}", e);
                Err(ApiError::InternalServerError)
            })?;

            return Ok(cache);
        } else {
            let result = PostEntity::find()
                .order_by_desc(PostColumn::CreatedAt)
                .limit(256)
                .all(self.db)
                .await
                .or_else(|err| match err {
                    DbErr::RecordNotFound(_) => Ok(vec![]),
                    _ => Err(ApiError::InternalServerError),
                })?;

            match serde_json::to_string(&result) {
                Ok(v) => {
                    _ = self
                        .cm
                        .set_ttl("post-recomendation".to_string(), v, 60)
                        .await
                        .or_else(|e| {
                            log::error!("Failed to cache post recomendations {}", e);
                            Err(())
                        });
                }
                Err(e) => {
                    log::error!("Failed to encode cached recomendation: {}", e);
                }
            }

            Ok(result)
        }
    }
}
