import logging
import os
from datetime import datetime

from src.frontend.ui_manager import MyApp
from src.shared import constants


def main():

    app = MyApp()
    app.geometry("800x400")
    app.show_page("StartPage")
    app.mainloop()


def setup_logger():
    # Create a logger
    logger = logging.getLogger()
    logger.setLevel(logging.DEBUG)

    # Create console handler and set level to debug
    ch = logging.StreamHandler()
    ch.setLevel(logging.DEBUG)

    # Ensure the logs directory exists
    log_dir = constants.LOGS_FOLDER_PATH

    os.makedirs(log_dir, exist_ok=True)

    # Create file handler with date and time in the filename
    log_filename = os.path.join(log_dir, f"app_{datetime.now().strftime('%Y%m%d_%H%M%S')}.log")
    fh = logging.FileHandler(log_filename)
    fh.setLevel(logging.DEBUG)

    # Create formatter
    formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s \n["%(filename)s:%(lineno)d"]')

    # Add formatter to handlers
    ch.setFormatter(formatter)
    fh.setFormatter(formatter)

    # Add handlers to logger
    logger.addHandler(ch)
    logger.addHandler(fh)


if __name__ == "__main__":
    # log to console
    setup_logger()
    main()
    # try:
    #     main()
    # except Exception as e:
    #     print(f"Error: {e}")
    #     sys.exit(1)