use crate::{
    error::ApiError,
    repository::auth::{AuthRepository, AuthService, UserJwtPayload},
};
use actix_web::{dev::Payload, web, Error as ActixError, FromRequest, HttpRequest};
use futures_util::{Future, FutureExt};
use std::pin::Pin;

pub struct AuthorizedUser {
    pub token: UserJwtPayload,
}

fn err_(e: ApiError) -> Pin<Box<dyn Future<Output = Result<AuthorizedUser, ActixError>>>> {
    async move { Err(ActixError::from(e)) }.boxed_local()
}

fn parse_auth_header(s: String) -> Result<String, ApiError> {
    let splited = s.split(" ");
    let splited = splited.collect::<Vec<&str>>();

    if splited.len() != 2 {
        return Err(ApiError::InvalidAuthHeaderFormat);
    }

    let method = splited[0];

    if method != "Bearer" {
        if method == "Signature" {
            Err(ApiError::SignatureAuthNotSupported)
        } else {
            Err(ApiError::InvalidAuthHeaderFormat)
        }
    } else {
        Ok(splited[1].to_string())
    }
}

impl FromRequest for AuthorizedUser {
    type Error = ActixError;

    type Future = Pin<Box<dyn Future<Output = Result<Self, Self::Error>>>>;

    fn from_request(req: &HttpRequest, _payload: &mut Payload) -> Self::Future {
        let auth_token = match req.headers().get("authorization") {
            Some(v) => match v.to_str() {
                Ok(v) => {
                    let result = parse_auth_header(v.to_string());

                    match result {
                        Ok(v) => v,
                        Err(e) => return err_(e),
                    }
                }
                Err(err) => {
                    log::warn!("Failed to decode request header {}", err);
                    return err_(ApiError::InternalServerError);
                }
            },
            None => {
                let auth_cookie = req.cookie("auth-token");

                match auth_cookie {
                    Some(v) => v.to_string(),
                    None => return err_(ApiError::AuthorizationRequired),
                }
            }
        };

        let auth_service = match req.app_data::<web::Data<AuthService>>() {
            Some(v) => v,
            None => {
                log::error!("Poorly configured actix AuthProvider dependency injection");
                return err_(ApiError::InternalServerError);
            }
        }
        .clone();

        async move {
            let token = auth_service
                .decode_token(auth_token)
                .await
                .or_else(|e| Err(ActixError::from(e)))?;

            Ok(Self { token })
        }
        .boxed_local()
    }
}
