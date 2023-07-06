use actix_web::{
    body::BoxBody,
    http::{header::ContentType, Method, StatusCode},
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

#[derive(Serialize, Deserialize)]
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
