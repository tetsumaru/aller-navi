"""Initial migration

Revision ID: ad5743981150
Revises:
Create Date: 2024-07-13 13:48:33.542186

"""

from typing import Sequence, Union

import sqlalchemy as sa

from alembic import op

# revision identifiers, used by Alembic.
revision: str = 'ad5743981150'
down_revision: Union[str, None] = None
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.create_table(
        'recipe_ingredients_map',
        sa.Column('id', sa.Integer, primary_key=True),
        sa.Column('recipe_name', sa.String(255), nullable=False),
        sa.Column('ingredient_name', sa.String(255), nullable=False),
        sa.UniqueConstraint(
            'recipe_name', 'ingredient_name', name='unique_recipe_ingredient'
        ),
    )

    op.create_table(
        'ingredient_allergen_map',
        sa.Column('id', sa.Integer, primary_key=True),
        sa.Column('ingredient_name', sa.String(255), nullable=False),
        sa.Column('allergen_name', sa.String(255), nullable=False),
        sa.UniqueConstraint(
            'ingredient_name',
            'allergen_name',
            name='unique_ingredient_allergen',
        ),
    )

    op.create_table(
        'user',
        sa.Column('id', sa.Integer, primary_key=True),
        sa.Column('name', sa.String(255), nullable=False),
        sa.Column('class', sa.String(255), nullable=True),
        sa.Column('age', sa.Integer, nullable=True),
    )

    op.create_table(
        'user_allergen_map',
        sa.Column('id', sa.Integer, primary_key=True),
        sa.Column(
            'user_id', sa.Integer, sa.ForeignKey('user.id'), nullable=False
        ),
        sa.Column('allergen', sa.String(255), nullable=False),
        sa.UniqueConstraint(
            'user_id', 'allergen', name='unique_user_allergen'
        ),
    )

    op.create_table(
        'monthly_menu',
        sa.Column('id', sa.Integer, primary_key=True),
        sa.Column('nursery_name', sa.String(255), nullable=False),
        sa.Column('class_name', sa.String(255), nullable=False),
        sa.Column('date', sa.Date, nullable=False),
        sa.Column('menu_name', sa.String(255), nullable=False),
    )

    op.create_table(
        'bento_day',
        sa.Column('id', sa.Integer, primary_key=True),
        sa.Column('nursery_name', sa.String(255), nullable=False),
        sa.Column('class_name', sa.String(255), nullable=False),
        sa.Column('date', sa.Date, nullable=False),
    )


def downgrade() -> None:
    op.drop_table('recipe_ingredients_map')
    op.drop_table('ingredient_allergen_map')
    op.drop_table('user')
    op.drop_table('user_allergen_map')
    op.drop_table('monthly_menu')
    op.drop_table('bento_day')
