from fastapi import FastAPI, Request
import json
from pymongo import MongoClient
import asyncio

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


async def revisar_y_eliminar_excesos():
    while True:
        query1 = {"worker1": "10.0.0.30"}
        documentos1 = list(collection.find(query1).sort("timestamp"))
        
        if len(documentos1) > 5:
            documentos_a_eliminar1 = len(documentos1) - 5
            for i in range(documentos_a_eliminar1):
                collection.delete_one({"_id": documentos1[i]["_id"]})
            print(f"Se han eliminado {documentos_a_eliminar1} documentos antiguos del worker1")
        
        query2 = {"worker2": "10.0.0.40"}
        documentos2 = list(collection.find(query2).sort("timestamp"))
        
        if len(documentos2) > 5:
            documentos_a_eliminar2 = len(documentos2) - 5
            for i in range(documentos_a_eliminar2):
                collection.delete_one({"_id": documentos2[i]["_id"]})
            print(f"Se han eliminado {documentos_a_eliminar2} documentos antiguos del worker2")


        query3 = {"worker3": "10.0.0.50"}
        documentos3 = list(collection.find(query3).sort("timestamp"))
        
        if len(documentos3) > 5:
            documentos_a_eliminar3 = len(documentos3) - 5
            for i in range(documentos_a_eliminar3):
                collection.delete_one({"_id": documentos3[i]["_id"]})
            print(f"Se han eliminado {documentos_a_eliminar3} documentos antiguos del worker2")

        await asyncio.sleep(5)  # Espera 5 segundos antes de revisar nuevamente

async def app_lifespan():
    await revisar_y_eliminar_excesos()

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="10.0.10.2", port=9898)