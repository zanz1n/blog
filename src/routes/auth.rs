use crate::{
    model::user::ApiUser,
    repository::{
        auth::AuthProvider,
        user::{CreateUserData, UserError, UserRepository},
    },
    utils::http::{serialize_response, DataBody},
};
use actix_web::{
    body::BoxBody,
    http::StatusCode,
    post,
    web::{Data, Json},
    CustomizeResponder, Responder,
};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct SignInRequestBody {
    email: String,
    password: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SignInResponseBody {
    token: String,
}

impl Responder for SignInResponseBody {
    type Body = BoxBody;

    fn respond_to(self, _: &actix_web::HttpRequest) -> actix_web::HttpResponse<Self::Body> {
        serialize_response(&self, StatusCode::OK)
    }
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SignUpResponseBody {
    data: ApiUser,
    message: String,
    token: Option<String>,
}

impl Responder for SignUpResponseBody {
    type Body = BoxBody;

    fn respond_to(self, _: &actix_web::HttpRequest) -> actix_web::HttpResponse<Self::Body> {
        serialize_response(&self, StatusCode::OK)
    }
}

#[post("/auth/signin")]
pub async fn signin(
    auth_provider: Data<AuthProvider>,
    body: Json<SignInRequestBody>,
) -> Result<SignInResponseBody, UserError> {
    let body = body.0;

    auth_provider
        .auth_user(body.email, body.password)
        .await
        .map(|token| SignInResponseBody { token })
}

#[post("/auth/signup")]
pub async fn signup(
    auth_provider: Data<AuthProvider>,
    user_repo: Data<UserRepository>,
    body: Json<CreateUserData>,
) -> Result<CustomizeResponder<SignUpResponseBody>, UserError> {
    let user = user_repo.create(body.0).await?;

    let token = auth_provider
        .generate_token(user.id.clone(), user.email.clone(), user.username.clone())
        .await;

    let token = match token {
        Ok(v) => Some(v),
        Err(_) => {
            return Ok(SignUpResponseBody {
                data: user.to_sendable(),
                message:
                    "User created successfully, but the auth token could not be generated due to an \
                    unexpected error of our part, try signin manually. This incident has been logged"
                        .to_string(),
                token: None,
            }
            .customize()
            .with_status(StatusCode::INTERNAL_SERVER_ERROR));
        }
    };

    Ok(SignUpResponseBody {
        data: user.to_sendable(),
        message: "User created successfully".to_string(),
        token,
    }
    .customize()
    .with_status(StatusCode::CREATED))
}
