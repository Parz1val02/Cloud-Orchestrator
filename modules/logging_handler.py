import logging
from pymongo import MongoClient
from datetime import datetime


class MongoDBHandler(logging.Handler):
    def __init__(self, db_name, collection_name, task_id):
        super().__init__()
        self.client = MongoClient("localhost", 27017)
        self.db = self.client[db_name]
        self.collection = self.db[collection_name]
        self.logs = []
        self.task_id = task_id

    def emit(self, record):
        log_entry = self.format(record)
        self.logs.append(
            {
                "task_id": self.task_id,
                "timestamp": datetime.now(),
                "message": log_entry,
                "level": record.levelname,
            }
        )

    def flush(self):
        if self.logs:
            self.collection.insert_many(self.logs)
            self.logs = []

    def close(self):
        self.flush()
        super().close()
