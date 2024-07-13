# app/main_view.py

import tkinter as tk

from app.controller import Controller


class MainView(tk.Tk):
    def __init__(self, controller):
        super().__init__()

        self.controller = controller
        self.title("Main View")
        self.geometry("800x600")

        self.label = tk.Label(self, text="Hello, Tkinter!")
        self.label.pack(pady=20)

        self.button = tk.Button(
            self, text="Click Me", command=self.controller.on_button_click
        )
        self.button.pack(pady=20)


if __name__ == "__main__":
    controller = Controller()
    app = MainView(controller)
    app.mainloop()
