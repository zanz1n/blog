use sea_orm_migration::{prelude::*, sea_query::extension::postgres::Type};

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    // SeaOrm migration code
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .create_type(
                Type::create()
                    .as_enum(UserRole::Enum)
                    .values([UserRole::Admin, UserRole::Common, UserRole::Publisher])
                    .to_owned(),
            )
            .await?;

        manager
            .create_table(
                Table::create()
                    .table(User::Table)
                    .if_not_exists()
                    .col(
                        ColumnDef::new(User::Id)
                            .string_len(18)
                            .not_null()
                            .primary_key(),
                    )
                    .col(ColumnDef::new(User::CreatedAt).timestamp().not_null())
                    .col(ColumnDef::new(User::UpdatedAt).timestamp().not_null())
                    .col(ColumnDef::new(User::Email).string_len(64).not_null())
                    .col(ColumnDef::new(User::Username).string_len(42).not_null())
                    .col(ColumnDef::new(User::Password).string_len(255).not_null())
                    .col(
                        ColumnDef::new(User::Role)
                            .enumeration(
                                UserRole::Enum,
                                [UserRole::Admin, UserRole::Common, UserRole::Publisher],
                            )
                            .not_null(),
                    )
                    .to_owned(),
            )
            .await?;

        manager
            .create_index(
                Index::create()
                    .table(User::Table)
                    .unique()
                    .col(User::Email)
                    .to_owned(),
            )
            .await?;

        Ok(())
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .drop_table(Table::drop().if_exists().table(User::Table).to_owned())
            .await?;

        manager
            .drop_type(Type::drop().if_exists().name(UserRole::Enum).to_owned())
            .await?;

        Ok(())
    }
}

#[derive(Iden)]
enum UserRole {
    #[iden = "userrole"]
    Enum,
    #[iden = "ADMIN"]
    Admin,
    #[iden = "PUBLISHER"]
    Publisher,
    #[iden = "COMMON"]
    Common,
}

#[derive(Iden)]
enum User {
    #[iden = "users"]
    Table,
    #[iden = "id"]
    Id,
    #[iden = "createdAt"]
    CreatedAt,
    #[iden = "updatedAt"]
    UpdatedAt,
    #[iden = "email"]
    Email,
    #[iden = "username"]
    Username,
    #[iden = "password"]
    Password,
    #[iden = "role"]
    Role,
}

// assert_eq!(User::Table.to_string(), "users");
// assert_eq!(User::Id.to_string(), "id");
// assert_eq!(User::CreatedAt.to_string(), "createdAt");
// assert_eq!(User::UpdatedAt.to_string(), "updatedAt");
// assert_eq!(User::Email.to_string(), "email");
// assert_eq!(User::Username.to_string(), "username");
// assert_eq!(User::Password.to_string(), "password");
