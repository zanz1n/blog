use crate::error::ErrorResponseBody;
use actix_web::{
    body::BoxBody,
    http::{header::ContentType, Method, StatusCode},
    web::JsonConfig,
    HttpResponse, Responder,
};
use serde::{Deserialize, Serialize};

pub const ENCODING_FAILED_BODY: &'static str =
    "{\"error\":\"The intended response body could not be encoded, this occurrence was logged\"}";

#[derive(Serialize, Deserialize)]
pub struct DataBody<T: Serialize> {
    data: T,
    message: String,
}

#[derive(Clone, Debug, Copy, Serialize, Deserialize)]
pub struct PathWithId<T> {
    pub id: T,
}

#[derive(Clone, Debug, Copy, Serialize, Deserialize)]
pub struct CursorLimitQueryParams<C, L> {
    pub limit: L,
    pub cursor: C,
}

impl<C, L> CursorLimitQueryParams<C, L>
where
    C: Clone,
    L: Clone,
{
    pub fn limit(&self) -> L {
        self.limit.clone()
    }

    pub fn cursor(&self) -> C {
        self.cursor.clone()
    }
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
        Err(e) => {
            log::warn!("Failed to encode response body: {}", e);
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
        .error_handler(|a, _| ErrorResponseBody::new(a, 4000).into())
}
