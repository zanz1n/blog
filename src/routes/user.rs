use crate::{
    error::ApiError,
    middlewares::auth::AuthorizedUser,
    model::user::{ApiUser, UserRole},
    repository::{
        auth::{AuthRepository, InvalidationReason},
        user::{CreateUserData, UpdateEmailData, UserRepository},
    },
    utils::http::{DataBody, PathWithId},
};
use actix_web::{
    delete, get, post, put,
    web::{Data, Json, Path},
};

#[get("/user/{id}")]
async fn get_by_id(
    user_repo: Data<dyn UserRepository>,
    path_params: Path<PathWithId<String>>,
) -> Result<ApiUser, ApiError> {
    let user = user_repo.get_by_id(path_params.id()).await?;

    Ok(user.to_sendable())
}

#[get("/user/self")]
async fn get_self(
    user_repo: Data<dyn UserRepository>,
    token: AuthorizedUser,
) -> Result<ApiUser, ApiError> {
    let user = user_repo.get_by_id(token.token.sub).await?;

    Ok(user.to_sendable())
}

#[post("/user")]
async fn create(
    user_repo: Data<dyn UserRepository>,
    data: Json<CreateUserData>,
) -> Result<ApiUser, ApiError> {
    let user = user_repo.create(data.0).await?;

    Ok(user.to_sendable())
}

#[put("/user/{id}/invalidate")]
async fn invalidate_user(
    auth_repository: Data<dyn AuthRepository>,
    token: AuthorizedUser,
    params: Path<PathWithId<String>>,
) -> Result<DataBody<Option<()>>, ApiError> {
    if token.token.sub != params.id && token.token.role != UserRole::Admin {
        Err(ApiError::DataMutationDenied)
    } else {
        auth_repository
            .add_invalidation(params.id(), InvalidationReason::UserRequest)
            .await?;

        Ok(DataBody::new(None, "User invalidation started"))
    }
}

#[put("/user/{id}/username")]
async fn update_username(
    user_repo: Data<dyn UserRepository>,
    data: Json<UpdateEmailData>,
    token: AuthorizedUser,
    params: Path<PathWithId<String>>,
) -> Result<ApiUser, ApiError> {
    if token.token.sub != params.id && token.token.role != UserRole::Admin {
        Err(ApiError::DataMutationDenied)
    } else {
        let user = user_repo.update_username(params.id(), data.0).await?;

        Ok(user.to_sendable())
    }
}

#[delete("/user/{id}")]
async fn delete_user(
    user_repo: Data<dyn UserRepository>,
    auth_service: Data<dyn AuthRepository>,
    token: AuthorizedUser,
    params: Path<PathWithId<String>>,
) -> Result<DataBody<Option<u8>>, ApiError> {
    if token.token.sub != params.id && token.token.role != UserRole::Admin {
        Err(ApiError::DataMutationDenied)
    } else {
        user_repo.delete(params.id()).await?;

        auth_service
            .add_invalidation(params.id(), InvalidationReason::UserDeleted)
            .await
            .unwrap_or_else(|e| log::error!("Failed to add user invalidation: {}", e));

        Ok(DataBody::new(None, "Deleted"))
    }
}
