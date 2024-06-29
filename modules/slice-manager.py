from flask import Flask, request, jsonify
from pymongo import MongoClient
from bson.objectid import ObjectId
import tests as openstack_driver

app = Flask(__name__)

# Configuración de la conexión a MongoDB
client = MongoClient("localhost", 27017)



def serialize_document(doc):
    # Crear una copia del documento original para evitar modificar mientras iteramos
    doc_copy = doc.copy()
    for key, value in doc_copy.items():
        if isinstance(value, ObjectId):
            # Convertir ObjectId a cadena de texto
            doc[key] = str(value)
            if key != "slice_id":
                doc["slice_id"] = doc.pop(key)
    return doc  # Función para serializar documentos MongoDB / necesario para ser enviados como respuesta del endpoint / working

def serialize_template(template):
    doc_copy = template.copy()
    for key, value in doc_copy.items():
        if isinstance(value, ObjectId):
            template[key] = str(value)
            if key != "template_id":
                template["template_id"] = template.pop(key)
    return template  

def obtenerTemplateById(template_id):
    db = client.cloud
    collection = db.templates
    template = collection.find_one({"_id": ObjectId(template_id)})
     # Eliminar el campo '_id' si existe
    if '_id' in template:
        del template['_id']

    return template
    

# Endpoint para crear una nueva plantilla / working
@app.route("/slices", methods=["POST"])
def crear_slice():
    new_slice_info = request.json
    new_slice = obtenerTemplateById(new_slice_info['template_id'])
    new_slice.update(new_slice_info)

    db = client.cloud
    collection = db.slices
    result = collection.insert_one(new_slice)
    if result.inserted_id:
        
        if new_slice['deployment_type']=='openstack':
            # implementa openstack .  
            
            user_name = request.headers["X-User-Username"]
            openstack_driver.openstackDeployment(new_slice,user_name)
            
            
            return jsonify(
                {
                    "msg": f"Slice with id {result.inserted_id} created successfully in OpenStack",
                    "result": "success",
                }
            )
        
        else:
            # implementa linux
            
            return jsonify(
                {
                    "msg": f"Slice with id {result.inserted_id} created successfully in Linux Cluster",
                    "result": "success",
                }
            )
        
       
    else:
        response = jsonify(
            {
                "result": "error",
                "msg": "Template not created due to error",
            }
        )
        error_code = 400
        return response, error_code


# Endpoint para listar todas las plantillas / working
@app.route("/slices", methods=["GET"])
def listar_slices():
    db = client.cloud
    collection = db.slices

    role = request.headers["X-User-Role"]
    if not role:
        response = jsonify({"result": "error", "msg": "User role is required"})
        error_code = 400
        return response, error_code
    else:
        fields = {
            "name": 1,
            "description": 1,
            "created_at": 1,
            "topology_type": 1,
            "availability_zone": 1,
            "deployment_type": 1,
            "internet": 1,
            "_id": 1,
        }
        if role == "user":
            user_id = request.headers["X-User-ID"]
            query = {"user_id": user_id}
            if not user_id:
                response = jsonify({"result": "error", "msg": "User id is required"})
                error_code = 400
                return response, error_code
            slices = [
                serialize_document(slice) for slice in collection.find(query, fields)
            ]
            if slices:
                return jsonify({"result": "success", "slices": slices})
            else:
                return jsonify(
                    {
                        "result": "success",
                        "msg": "No available slices to display",
                    }
                )

        elif role == "administrator":
            query = {}
            slices = [
                serialize_document(slice) for slice in collection.find(query, fields)
            ]
            if slices:
                return jsonify({"result": "success", "slices": slices})
            else:
                return jsonify(
                    {
                        "result": "success",
                        "msg": "No available slices to display",
                    }
                )
        else:
            response = jsonify({"result": "error", "msg": "Invalid role"})
            error_code = 400
            return response, error_code


if __name__ == "__main__":
    app.run(debug=True, port=9999)
