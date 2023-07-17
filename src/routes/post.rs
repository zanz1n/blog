use crate::{
    error::ApiError,
    middlewares::auth::AuthorizedUser,
    model::post::Model as PostModel,
    repository::post::{CreatePostData, PostRepository, PostService},
    utils::http::{DataBody, LimitQueryParams, PathWithId},
};
use actix_web::{
    get, post,
    web::{Data, Json, Path, Query},
};

#[post("/post")]
async fn create_post(
    post_repo: Data<PostService>,
    data: Json<CreatePostData>,
    auth_data: AuthorizedUser,
) -> Result<PostModel, ApiError> {
    post_repo.create(auth_data.token.sub, data.0).await
}

#[get("/posts")]
async fn get_posts_recomendation(
    post_repo: Data<PostService>,
    query: Query<LimitQueryParams<Option<usize>>>,
) -> Result<DataBody<Vec<PostModel>>, ApiError> {
    let result = post_repo
        .get_recomendation(query.limit.unwrap_or_else(|| 200))
        .await?;

    Ok(DataBody::new(result, "Success"))
}

#[get("/post/{id}")]
async fn get_post_by_id(
    post_repo: Data<PostService>,
    params: Path<PathWithId<String>>,
) -> Result<PostModel, ApiError> {
    post_repo.get_by_id(params.id()).await
}

#[get("/user/{id}/posts")]
async fn get_user_posts(
    post_repo: Data<PostService>,
    params: Path<PathWithId<String>>,
    query: Query<LimitQueryParams<Option<u64>>>,
) -> Result<DataBody<Vec<PostModel>>, ApiError> {
    let result = post_repo
        .get_user_posts(params.id(), query.limit.unwrap_or_else(|| 200))
        .await?;

    Ok(DataBody::new(result, "Success"))
}
