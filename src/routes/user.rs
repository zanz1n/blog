use actix_web::{
    get, post,
    web::{Data, Json, Path},
};

use crate::{
    model::{user::ApiUser},
    repository::user::{CreateUserData, UserError, UserRepository},
    utils::http::PathWithId,
};

#[get("/user/{id}")]
async fn get_by_id(
    user_repo: Data<UserRepository>,
    path_params: Path<PathWithId<String>>,
) -> Result<ApiUser, UserError> {
    let user = user_repo.get_by_id(path_params.id()).await?;

    Ok(user.to_sendable())
}

#[post("/user")]
async fn create(
    user_repo: Data<UserRepository>,
    data: Json<CreateUserData>,
) -> Result<ApiUser, UserError> {
    let user = user_repo.create(data.0).await?;

    Ok(user.to_sendable())
}
