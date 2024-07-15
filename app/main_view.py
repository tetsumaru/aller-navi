import csv
import tkinter as tk
from tkinter import filedialog, messagebox

from sqlalchemy.dialects.sqlite import insert

from app.infrastructure.database.sqlite_handler import (
    recipe_ingredients_map,
    session,
)


def load_csv_to_db(file_path: str) -> None:
    """Load data from a CSV file into the database."""
    with open(file_path, newline='', encoding='utf-8') as csvfile:
        reader = csv.DictReader(csvfile)
        for row in reader:
            insert_stmt = insert(recipe_ingredients_map).values(
                recipe_name=row['recipe_name'],
                ingredient_name=row['ingredient_name'],
            )
            session.execute(insert_stmt)
        session.commit()


def select_file() -> None:
    """Open a file dialog to select a CSV file and load its data into the database."""
    file_path = filedialog.askopenfilename(
        title="Select CSV file", filetypes=[("CSV files", "*.csv")]
    )
    if file_path:
        try:
            load_csv_to_db(file_path)
            messagebox.showinfo("Success", "Data loaded successfully!")
        except Exception as e:
            messagebox.showerror("Error", f"Failed to load data: {e}")


# Tkinter GUIの設定
root = tk.Tk()
root.title("CSV Importer")

# ファイル選択ボタン
button = tk.Button(root, text="Select CSV File", command=select_file)
button.pack(pady=20)

# メインループ
root.mainloop()
