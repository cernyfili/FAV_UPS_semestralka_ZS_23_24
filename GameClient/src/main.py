import logging

from frontend.ui_manager import MyApp


def main():

    app = MyApp()
    app.geometry("800x400")
    app.show_page("StartPage")
    app.mainloop()


if __name__ == "__main__":
    # log to console
    logging.basicConfig(
        level=logging.DEBUG,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s \n["%(filename)s:%(lineno)d"]'
    )
    main()
    # try:
    #     main()
    # except Exception as e:
    #     print(f"Error: {e}")
    #     sys.exit(1)