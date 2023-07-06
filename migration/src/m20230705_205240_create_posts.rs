use sea_orm_migration::prelude::*;

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    // SeaOrm migration code
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .create_table(
                Table::create()
                    .table(Post::Table)
                    .if_not_exists()
                    .col(
                        ColumnDef::new(Post::Id)
                            .string_len(24)
                            .primary_key()
                            .not_null(),
                    )
                    .col(ColumnDef::new(Post::CreatedAt).timestamp().not_null())
                    .col(ColumnDef::new(Post::UpdatedAt).timestamp().not_null())
                    .col(ColumnDef::new(Post::Slug).string_len(64).not_null())
                    .col(ColumnDef::new(Post::Content).text().not_null())
                    .col(ColumnDef::new(Post::ThumbImage).string_len(128).null())
                    .col(ColumnDef::new(Post::UserId).string_len(18).not_null())
                    .to_owned(),
            )
            .await?;

        manager
            .create_index(
                Index::create()
                    .table(Post::Table)
                    .unique()
                    .col(Post::UserId)
                    .to_owned(),
            )
            .await?;

        manager
            .create_index(
                Index::create()
                    .table(Post::Table)
                    .col(Post::Slug)
                    .to_owned(),
            )
            .await?;

        Ok(())
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .drop_table(Table::drop().table(Post::Table).to_owned())
            .await
    }
}

#[derive(Iden)]
enum Post {
    #[iden = "posts"]
    Table,
    #[iden = "id"]
    Id,
    #[iden = "createdAt"]
    CreatedAt,
    #[iden = "updatedAt"]
    UpdatedAt,
    #[iden = "slug"]
    Slug,
    #[iden = "content"]
    Content,
    #[iden = "thumbImageKey"]
    ThumbImage,
    #[iden = "userId"]
    UserId,
}

// assert_eq!(Post::Table.to_string(), "posts");
// assert_eq!(Post::Id.to_string(), "id");
// assert_eq!(Post::CreatedAt.to_string(), "createdAt");
// assert_eq!(Post::UpdatedAt.to_string(), "updatedAt");
// assert_eq!(Post::Slug.to_string(), "slug");
// assert_eq!(Post::Content.to_string(), "content");
// assert_eq!(Post::ThumbImage.to_string(), "thumbImageKey");
// assert_eq!(Post::UserId.to_string(), "userId");
