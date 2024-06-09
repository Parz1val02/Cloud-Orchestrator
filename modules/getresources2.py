from fastapi import FastAPI, Request
import json
from pymongo import MongoClient


app = FastAPI()



client = MongoClient('localhost', 27017)
db = client['cloud']
collection = db['resources']




@app.post("/data")
async def receive_data(request: Request):
    data = await request.json()
    print(f"Received data: {data}")
    result = collection.insert_one(data)
    return {"status": "success"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="10.0.10.2", port=9898)