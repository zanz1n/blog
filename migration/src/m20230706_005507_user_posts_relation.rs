use sea_orm_migration::prelude::*;

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .create_foreign_key(
                ForeignKey::create()
                    .name(POST_USER_FOREIGN_KEY)
                    .from_tbl(Post::Table)
                    .to_tbl(User::Table)
                    .from_col(Post::UserId)
                    .to_col(User::Id)
                    .on_update(ForeignKeyAction::Cascade)
                    .on_delete(ForeignKeyAction::Restrict)
                    .to_owned(),
            )
            .await
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .drop_foreign_key(
                ForeignKey::drop()
                    .name(POST_USER_FOREIGN_KEY)
                    .table(Post::Table)
                    .to_owned(),
            )
            .await
    }
}

const POST_USER_FOREIGN_KEY: &'static str = "post_userId_fkey";

#[derive(Iden)]
enum User {
    #[iden = "users"]
    Table,
    #[iden = "id"]
    Id,
}

#[derive(Iden)]
enum Post {
    #[iden = "posts"]
    Table,
    #[iden = "userId"]
    UserId,
}
