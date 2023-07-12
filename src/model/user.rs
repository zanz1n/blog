use crate::utils::http::DataBody;
use actix_web::{body::BoxBody, HttpResponse};
use sea_orm::prelude::*;
use serde::{Deserialize, Serialize};

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
    pub role: UserRole,
}

#[derive(Clone, Debug, PartialEq, EnumIter, DeriveActiveEnum, Serialize, Deserialize)]
#[sea_orm(rs_type = "String", db_type = "Enum", enum_name = "userrole")]
pub enum UserRole {
    #[sea_orm(string_value = "COMMON")]
    #[serde(rename = "COMMON")]
    Common,
    #[sea_orm(string_value = "ADMIN")]
    #[serde(rename = "ADMIN")]
    Admin,
    #[sea_orm(string_value = "PUBLISHER")]
    #[serde(rename = "PUBLISHER")]
    Publisher,
}

impl ToString for UserRole {
    fn to_string(&self) -> String {
        match self {
            UserRole::Admin => "ADMIN",
            UserRole::Common => "COMMON",
            UserRole::Publisher => "PUBLISHER",
        }
        .to_owned()
    }
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
    pub role: UserRole,
}

impl Model {
    pub fn to_sendable(self) -> ApiUser {
        ApiUser {
            id: self.id,
            created_at: self.created_at,
            updated_at: self.updated_at,
            email: self.email,
            username: self.username,
            role: self.role,
        }
    }
}

impl actix_web::Responder for ApiUser {
    type Body = BoxBody;

    fn respond_to(self, req: &actix_web::HttpRequest) -> HttpResponse<Self::Body> {
        DataBody::new(self, "Success").respond_to(req)
    }
}
