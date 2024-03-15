import logging
import os

class Logger:
    
    def get_logger(self):
        LOGLEVEL = os.environ.get('LOGLEVEL', 'INFO').upper()
        SET_LEVEL = 'logger.setLevel(logging.' + LOGLEVEL + ')'
        SET_CH_LEVEL = 'ch.setLevel(logging.' + LOGLEVEL + ')'
        logger = logging.getLogger(self)
        exec(SET_LEVEL)
        ch = logging.StreamHandler()
        exec(SET_CH_LEVEL)
        formatter = logging.Formatter('%(asctime)s  %(levelname)s:     %(message)s')
        ch.setFormatter(formatter)
        logger.addHandler(ch)
        return logger