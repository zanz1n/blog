use actix_web::{body::BoxBody, HttpResponse};
use sea_orm::prelude::*;
use serde::{Deserialize, Serialize};

use crate::utils::http::DataBody;

#[derive(Clone, Debug, PartialEq, DeriveEntityModel)]
#[sea_orm(table_name = "users", schema_name = "public")]
pub struct Model {
    #[sea_orm(primary_key, column_type = "String(Some(18))")]
    pub id: String,
    #[sea_orm(column_name = "createdAt")]
    pub created_at: DateTime,
    #[sea_orm(column_name = "updatedAt")]
    pub updated_at: DateTime,
    #[sea_orm(column_type = "String(Some(64))", unique, indexed)]
    pub email: String,
    #[sea_orm(column_type = "String(Some(42))")]
    pub username: String,
    #[sea_orm(column_type = "String(Some(255))")]
    pub password: String,
}

#[derive(Copy, Clone, Debug, EnumIter, DeriveRelation)]
pub enum Relation {
    #[sea_orm(has_many = "super::post::Entity")]
    Post,
}

impl Related<super::post::Entity> for Entity {
    fn to() -> RelationDef {
        Relation::Post.def()
    }
}

impl ActiveModelBehavior for ActiveModel {}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct ApiUser {
    pub id: String,
    #[serde(rename = "createdAt")]
    pub created_at: DateTime,
    #[serde(rename = "updatedAt")]
    pub updated_at: DateTime,
    pub email: String,
    pub username: String,
}

impl Model {
    pub fn to_sendable(self) -> ApiUser {
        ApiUser {
            id: self.id,
            created_at: self.created_at,
            updated_at: self.updated_at,
            email: self.email,
            username: self.username,
        }
    }
}

impl actix_web::Responder for ApiUser {
    type Body = BoxBody;

    fn respond_to(self, req: &actix_web::HttpRequest) -> HttpResponse<Self::Body> {
        DataBody::new(self, "Success").respond_to(req)
    }
}
