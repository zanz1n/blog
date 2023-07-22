use crate::{
    error::ApiError,
    middlewares::auth::AuthorizedUser,
    model::post::{Model as PostModel, PostWithUser},
    repository::post::{CreatePostData, PostRepository},
    utils::{
        html::{self, HeadingNode},
        http::{CursorLimitQueryParams, DataBody, PathWithId},
    },
};
use actix_web::{
    get, post,
    web::{Data, Json, Path, Query},
};

#[post("/post")]
async fn create_post(
    post_repo: Data<dyn PostRepository>,
    data: Json<CreatePostData>,
    auth_data: AuthorizedUser,
) -> Result<PostModel, ApiError> {
    post_repo.create(auth_data.token.sub, data.0).await
}

#[get("/posts")]
async fn get_posts_recomendation(
    post_repo: Data<dyn PostRepository>,
    query: Query<CursorLimitQueryParams<Option<usize>, Option<usize>>>,
) -> Result<DataBody<Vec<PostModel>>, ApiError> {
    let result = post_repo
        .get_recomendation(
            query.limit.unwrap_or_else(|| 200),
            query.cursor.unwrap_or_else(|| 0),
        )
        .await?;

    Ok(DataBody::new(result, "Success"))
}

#[get("/post/{id}/headings")]
async fn get_post_headings(
    post_repo: Data<dyn PostRepository>,
    params: Path<PathWithId<String>>,
) -> Result<DataBody<Vec<HeadingNode>>, ApiError> {
    let post = post_repo.get_by_id(params.id()).await?;

    let headings = html::get_headings(post.content.as_str());

    Ok(DataBody::new(headings, "Success"))
}

#[get("/post/{id}")]
async fn get_post_by_id(
    post_repo: Data<dyn PostRepository>,
    params: Path<PathWithId<String>>,
) -> Result<PostWithUser, ApiError> {
    let post = post_repo.get_by_id(params.id()).await?;

    Ok(post)
}

#[get("/user/{id}/posts")]
async fn get_user_posts(
    post_repo: Data<dyn PostRepository>,
    params: Path<PathWithId<String>>,
    query: Query<CursorLimitQueryParams<Option<u64>, Option<u64>>>,
) -> Result<DataBody<Vec<PostModel>>, ApiError> {
    let result = post_repo
        .get_user_posts(
            params.id(),
            query.limit.unwrap_or_else(|| 200),
            query.cursor.unwrap_or_else(|| 0),
        )
        .await?;

    Ok(DataBody::new(result, "Success"))
}
