# app/widgets.py

import tkinter as tk


class CustomButton(tk.Button):
    def __init__(self, master=None, **kwargs):
        super().__init__(master, **kwargs)
        self.config(bg="lightblue", fg="darkblue")
