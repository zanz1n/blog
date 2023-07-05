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
        let body_str: String;
        let status_code: StatusCode;

        match serde_json::to_string(&self) {
            Ok(enc) => {
                match req.method().into() {
                    Method::POST => {
                        status_code = StatusCode::CREATED;
                    }
                    _ => {
                        status_code = StatusCode::OK;
                    }
                }
                body_str = enc;
            }
            Err(_) => {
                status_code = StatusCode::INTERNAL_SERVER_ERROR;
                body_str = ENCODING_FAILED_BODY.to_string();
            }
        };

        HttpResponse::build(status_code)
            .content_type(ContentType::json())
            .body(body_str)
    }
}
