use crate::{
    model::user::ApiUser,
    repository::user::{CreateUserData, UpdateEmailData, UserError, UserRepository},
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

#[put("/user/{id}")]
async fn update_user(
    user_repo: Data<UserRepository>,
    data: Json<UpdateEmailData>,
    path_params: Path<PathWithId<String>>,
) -> Result<ApiUser, UserError> {
    let user = user_repo
        .update_username(path_params.id.clone(), data.0)
        .await?;

    Ok(user.to_sendable())
}

#[delete("/user/{id}")]
async fn delete_user(
    user_repo: Data<UserRepository>,
    path_params: Path<PathWithId<String>>,
) -> Result<DataBody<Option<u8>>, UserError> {
    user_repo.delete(path_params.id.clone()).await?;

    Ok(DataBody::new(None, "Deleted"))
}
