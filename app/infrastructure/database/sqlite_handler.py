from sqlalchemy import MetaData, Table, create_engine
from sqlalchemy.orm import sessionmaker

DATABASE_URL = "sqlite:///alembic/allernavi.db"

# SQLAlchemyのエンジンとセッションを設定
engine = create_engine(DATABASE_URL)
Session = sessionmaker(bind=engine)
session = Session()

# メタデータとテーブルの定義
metadata = MetaData()
recipe_ingredients_map = Table(
    'recipe_ingredients_map', metadata, autoload_with=engine
)
