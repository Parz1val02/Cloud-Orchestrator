from flask import Flask, request, jsonify
from pymongo import MongoClient
from bson.objectid import ObjectId

app = Flask(__name__)

# Configuración de la conexión a MongoDB
client = MongoClient("localhost", 27017)


# Función para serializar documentos MongoDB / necesario para ser enviados como respuesta del endpoint / working
def serialize_document(doc):
    for key, value in doc.items():
        if isinstance(value, ObjectId):
            #  Convertir ObjectId a cadena de texto
            doc[key] = str(value)
            doc["template_id"] = doc.pop(key)
    return doc


# Endpoint para listar todas las plantillas / working
@app.route("/templates", methods=["GET"])
def listar_plantillas():
    db = client.cloud
    collection = db.templates

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
            "_id": 1,
        }
        if role == "user":
            user_id = request.headers["X-User-ID"]
            query = {"user_id": user_id}
            if not user_id:
                response = jsonify({"result": "error", "msg": "User id is required"})
                error_code = 400
                return response, error_code
            templates = [
                serialize_document(template)
                for template in collection.find(query, fields)
            ]
            if templates:
                return jsonify({"result": "success", "templates": templates})
            else:
                return jsonify(
                    {
                        "result": "success",
                        "msg": "No available templates to display",
                    }
                )

        elif role == "administrator":
            query = {}
            templates = [
                serialize_document(template)
                for template in collection.find(query, fields)
            ]
            if templates:
                return jsonify({"result": "success", "templates": templates})
            else:
                return jsonify(
                    {
                        "result": "success",
                        "msg": "No available templates to display",
                    }
                )
        else:
            response = jsonify({"result": "error", "msg": "Invalid role"})
            error_code = 400
            return response, error_code


# Endpoint para buscar una plantilla por ID / working
@app.route("/templates/<string:template_id>", methods=["GET"])
def buscar_plantilla(template_id):
    db = client.cloud
    collection = db.templates
    try:
        template = collection.find_one({"_id": ObjectId(template_id)})
        if template:
            template_serialized = serialize_document(template)
            response = jsonify({"result": "success", "template": template_serialized})
            return response
        else:
            response = jsonify(
                {
                    "result": "error",
                    "msg": f"Template with template id {template_id} not found",
                }
            )
            error_code = 404
            return response, error_code
    except:
        response = jsonify(
            {"result": "error", "msg": f"Invalid template id: {template_id}"}
        )
        error_code = 400
        return response, error_code


# Endpoint para crear una nueva plantilla / working
@app.route("/templates", methods=["POST"])
def crear_plantilla():
    db = client.cloud
    collection = db.templates
    new_template = request.json
    result = collection.insert_one(new_template)
    if result.inserted_id:
        return jsonify(
            {
                "msg": f"Template with id {result.inserted_id} created successfully",
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


# Endpoint para editar una plantilla por ID / working
@app.route("/templates/<string:template_id>", methods=["PUT"])
def editar_plantilla(template_id):
    db = client.cloud
    collection = db.templates
    plantilla_actualizada = request.json

    try:
        result = collection.update_one(
            {"_id": ObjectId(template_id)}, {"$set": plantilla_actualizada}
        )
        if result.modified_count == 1:
            return jsonify(
                {
                    "result": "success",
                    "msg": f"Template with template id {template_id} updated successfully",
                }
            )
        else:
            response = jsonify(
                {
                    "result": "error",
                    "msg": f"Template with template id {template_id} not updated due to error",
                }
            )
            error_code = 400
            return response
    except:
        response = jsonify(
            {"result": "error", "msg": f"Invalid template id: {template_id}"}
        )
        error_code = 400
        return response, error_code


# Endpoint para eliminar una plantilla por ID / working
@app.route("/templates/<string:template_id>", methods=["DELETE"])
def eliminar_plantilla(template_id):
    db = client.cloud
    collection = db.templates
    try:
        result = collection.delete_one({"_id": ObjectId(template_id)})
        if result.deleted_count == 1:
            return jsonify(
                {
                    "result": "success",
                    "msg": f"Template with template id {template_id} deleted successfully",
                }
            )
        else:
            response = jsonify(
                {
                    "result": "error",
                    "msg": f"Template with template id {template_id} not deleted due to error",
                }
            )
            error_code = 404
            return response, error_code
    except:
        response = jsonify(
            {"result": "error", "msg": f"Invalid template id: {template_id} "}
        )
        error_code = 400
        return response, error_code


"""
# Endpoint para retornar link de topology graph / verify later
@app.route("/templates/<string:template_id>/graph", methods=["GET"])
def graph_plantilla(template_id):
    ### ACCIÓN PARA RECOGER EL TEMPLATE USANDO EL TEMPLATE_ID. OBTENER SOLO EL OBJETO TOPOLOGY (SIN ESPECIFICACIONES)

    ###
    link = "link"
    response = jsonify(
        {
            "result": "success",
            "topology_link": link,
            "msg": f"URL de topologia de plantilla con template_id {template_id} obtenida correctamente",
        }
    )
"""


def serialize_document_not_template(doc):
    for key, value in doc.items():
        if isinstance(value, ObjectId):
            #  Convertir ObjectId a cadena de texto
            doc[key] = str(value)
            doc["id"] = doc.pop(key)
    return doc


# Endpoint para listar todos los flavors
@app.route("/templates/flavors", methods=["GET"])
def listar_sabores():
    db = client.cloud
    collection = db.flavors
    flavors = [serialize_document_not_template(flavor) for flavor in collection.find()]
    if flavors:
        return jsonify({"result": "success", "flavors": flavors})
    else:
        return jsonify({"result": "success", "msg": "No available flavors to display"})


# Endpoint para listar todos los images
@app.route("/templates/images", methods=["GET"])
def listar_images():
    db = client.cloud
    collection = db.images
    images = [serialize_document_not_template(image) for image in collection.find()]
    if images:
        return jsonify({"result": "success", "images": images})
    else:
        return jsonify({"result": "success", "msg": "No available images to display"})


if __name__ == "__main__":
    app.run(debug=True)
