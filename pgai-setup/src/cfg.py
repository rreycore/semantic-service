from pydantic_settings import BaseSettings, SettingsConfigDict


class Cfg(BaseSettings):
    postgres_host: str
    postgres_user: str
    postgres_password: str
    postgres_db: str
    postgres_port: int
    postgres_sslmode: str

    model_config = SettingsConfigDict(env_file=".env")

    def get_db_url(self) -> str:
        return (
            f"host={self.postgres_host} "
            f"port={self.postgres_port} "
            f"user={self.postgres_user} "
            f"password={self.postgres_password} "
            f"dbname={self.postgres_db} "
            f"sslmode={self.postgres_sslmode}"
        )


cfg = Cfg.model_validate({})
