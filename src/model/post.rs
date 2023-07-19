use super::user::ApiUser;
use crate::utils::http::DataBody;
use actix_web::{body::BoxBody, HttpRequest, HttpResponse, Responder};
use sea_orm::prelude::*;
use serde::{Deserialize, Serialize};

#[derive(Clone, Debug, PartialEq, DeriveEntityModel, Serialize, Deserialize)]
#[sea_orm(table_name = "posts", schema_name = "public")]
pub struct Model {
    #[sea_orm(primary_key, column_type = "String(Some(24))")]
    pub id: String,
    #[sea_orm(column_name = "createdAt")]
    #[serde(rename = "createdAt")]
    pub created_at: DateTime,
    #[sea_orm(column_name = "updatedAt")]
    #[serde(rename = "updatedAt")]
    pub updated_at: DateTime,
    #[sea_orm(column_type = "String(Some(64))", unique, indexed)]
    pub slug: String,
    #[sea_orm(column_type = "Text")]
    pub content: String,
    #[sea_orm(column_name = "thumbImageKey", column_type = "String(Some(128))")]
    #[serde(rename = "thumbImage")]
    pub thumb_image: Option<String>,
    #[sea_orm(column_name = "userId", column_type = "String(Some(18))", indexed)]
    #[serde(rename = "userId")]
    pub user_id: String,
}

impl Model {
    pub fn with_user(self, user: Option<ApiUser>) -> PostWithUser {
        PostWithUser::new(self, user)
    }
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct PostWithUser {
    pub id: String,
    #[serde(rename = "createdAt")]
    pub created_at: DateTime,
    #[serde(rename = "updatedAt")]
    pub updated_at: DateTime,
    pub slug: String,
    pub content: String,
    #[serde(rename = "thumbImage")]
    pub thumb_image: Option<String>,
    pub user: Option<ApiUser>,
}

impl PostWithUser {
    pub fn new(model: Model, user: Option<ApiUser>) -> Self {
        Self {
            id: model.id,
            created_at: model.created_at,
            updated_at: model.updated_at,
            slug: model.slug,
            content: model.content,
            thumb_image: model.thumb_image,
            user,
        }
    }
}

impl actix_web::Responder for PostWithUser {
    type Body = BoxBody;

    fn respond_to(self, req: &HttpRequest) -> HttpResponse<Self::Body> {
        DataBody::new(self, "Success").respond_to(req)
    }
}

#[derive(Copy, Clone, Debug, EnumIter, DeriveRelation)]
pub enum Relation {
    #[sea_orm(
        belongs_to = "super::user::Entity",
        from = "Column::UserId",
        to = "super::user::Column::Id"
    )]
    User,
}

impl Related<super::user::Entity> for Entity {
    fn to() -> RelationDef {
        Relation::User.def()
    }
}

impl ActiveModelBehavior for ActiveModel {}

impl Responder for Model {
    type Body = BoxBody;

    fn respond_to(self, req: &HttpRequest) -> HttpResponse<Self::Body> {
        DataBody::new(self, "Success").respond_to(req)
    }
}
