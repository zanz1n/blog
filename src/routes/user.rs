use crate::{
    error::ApiError,
    middlewares::auth::AuthorizedUser,
    model::user::{ApiUser, UserRole},
    repository::user::{CreateUserData, UpdateEmailData, UserRepository},
    utils::http::{DataBody, PathWithId},
};
use actix_web::{
    delete, get, post, put,
    web::{Data, Json, Path},
};

#[get("/user/{id}")]
async fn get_by_id(
    user_repo: Data<UserRepository>,
    path_params: Path<PathWithId<String>>,
) -> Result<ApiUser, ApiError> {
    let user = user_repo.get_by_id(path_params.id()).await?;

    Ok(user.to_sendable())
}

#[get("/user/self")]
async fn get_self(
    user_repo: Data<UserRepository>,
    token: AuthorizedUser,
) -> Result<ApiUser, ApiError> {
    let user = user_repo.get_by_id(token.token.sub).await?;

    Ok(user.to_sendable())
}

#[post("/user")]
async fn create(
    user_repo: Data<UserRepository>,
    data: Json<CreateUserData>,
) -> Result<ApiUser, ApiError> {
    let user = user_repo.create(data.0).await?;

    Ok(user.to_sendable())
}

#[put("/user/{id}")]
async fn update_user(
    user_repo: Data<UserRepository>,
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
    user_repo: Data<UserRepository>,
    token: AuthorizedUser,
    params: Path<PathWithId<String>>,
) -> Result<DataBody<Option<u8>>, ApiError> {
    if token.token.sub != params.id && token.token.role != UserRole::Admin {
        Err(ApiError::DataMutationDenied)
    } else {
        user_repo.delete(params.id()).await?;

        Ok(DataBody::new(None, "Deleted"))
    }
}
