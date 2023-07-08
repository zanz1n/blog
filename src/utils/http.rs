use actix_web::{
    body::BoxBody,
    http::{header::ContentType, Method, StatusCode},
    web::JsonConfig,
    HttpResponse, Responder, ResponseError,
};
use serde::{Deserialize, Serialize};
use std::fmt::Display;

pub const ENCODING_FAILED_BODY: &'static str =
    "{\"error\":\"The intended response body could not be encoded, this occurrence was logged\"}";

#[derive(Serialize, Deserialize)]
pub struct DataBody<T: Serialize> {
    data: T,
    message: String,
}

#[derive(Clone, Copy, Serialize, Deserialize)]
pub struct PathWithId<T: Clone> {
    pub id: T,
}

impl<T: Clone> PathWithId<T> {
    pub fn id(&self) -> T {
        self.id.clone()
    }
}

pub fn serialize_response<T>(data: &T, status_code: StatusCode) -> HttpResponse<BoxBody>
where
    T: Serialize,
{
    let body_str: String;

    let status: StatusCode;

    match serde_json::to_string(data) {
        Ok(enc) => {
            status = status_code;
            body_str = enc;
        }
        Err(_) => {
            log::warn!("Failed to encode response body");
            status = StatusCode::INTERNAL_SERVER_ERROR;
            body_str = ENCODING_FAILED_BODY.to_string();
        }
    };

    HttpResponse::build(status)
        .content_type(ContentType::json())
        .body(body_str)
}

#[inline(always)]
fn code_from_method(method: Method) -> StatusCode {
    match method {
        Method::POST => StatusCode::CREATED,
        _ => StatusCode::OK,
    }
}

#[derive(Default, Debug, Serialize, Deserialize)]
pub struct ErrorResponseBody {
    error: String,
}

impl ErrorResponseBody {
    pub fn from<T: ToString>(el: T) -> Self {
        Self {
            error: el.to_string(),
        }
    }
}

impl Display for ErrorResponseBody {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(format!("{:?}", self).as_str())
    }
}

impl ResponseError for ErrorResponseBody {
    fn status_code(&self) -> StatusCode {
        StatusCode::BAD_REQUEST
    }

    fn error_response(&self) -> HttpResponse<BoxBody> {
        serialize_response(&self, self.status_code())
    }
}

impl<T: Serialize> DataBody<T> {
    pub fn new(data: T, msg: &str) -> DataBody<T> {
        DataBody {
            data,
            message: msg.to_string(),
        }
    }
}

impl<T: Serialize> Responder for DataBody<T> {
    type Body = BoxBody;

    fn respond_to(self, req: &actix_web::HttpRequest) -> actix_web::HttpResponse<Self::Body> {
        serialize_response(&self, code_from_method(req.method().into()))
    }
}

#[inline]
pub fn app_json_error_handler() -> JsonConfig {
    JsonConfig::default()
        .limit(4096) // 4kb
        .content_type_required(false)
        .content_type(|m| {
            (m.type_() == "text" && m.subtype() == "plain")
                || (m.type_() == "application" && m.subtype() == "json")
        })
        .error_handler(|a, _| ErrorResponseBody::from(a).into())
}
