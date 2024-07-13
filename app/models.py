# app/models.py

from sqlalchemy import Column, Integer, String, create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker

Base = declarative_base()


class ExampleModel(Base):
    __tablename__ = 'examples'

    id = Column(Integer, primary_key=True)
    name = Column(String, nullable=False)


# データベースの設定
DATABASE_URL = "sqlite:///example.db"

engine = create_engine(DATABASE_URL)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

# テーブルを作成
Base.metadata.create_all(bind=engine)
