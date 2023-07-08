use crate::utils::http::DataBody;
use actix_web::body::BoxBody;
use sea_orm::prelude::*;
use serde::{Deserialize, Serialize};

#[derive(Clone, Debug, PartialEq, DeriveEntityModel, Serialize, Deserialize)]
#[sea_orm(table_name = "users", schema_name = "public")]
pub struct Model {
    #[sea_orm(primary_key, column_type = "String(Some(24))")]
    pub id: String,
    #[sea_orm(column_name = "createdAt")]
    pub created_at: DateTime,
    #[sea_orm(column_name = "updatedAt")]
    pub updated_at: DateTime,
    #[sea_orm(column_type = "String(Some(64))", unique, indexed)]
    pub slug: String,
    #[sea_orm(column_type = "Text")]
    pub content: String,
    #[sea_orm(column_name = "thumbImageKey", column_type = "String(Some(128))")]
    pub thumb_image: String,
    #[sea_orm(column_name = "userId", column_type = "String(Some(18))", indexed)]
    pub user_id: String,
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

impl actix_web::Responder for Model {
    type Body = BoxBody;

    fn respond_to(self, req: &actix_web::HttpRequest) -> actix_web::HttpResponse<Self::Body> {
        DataBody::new(self, "Success").respond_to(req)
    }
}
