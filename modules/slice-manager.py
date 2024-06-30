from flask import Flask, request, jsonify
from celery.result import AsyncResult
from pymongo import MongoClient
from bson.objectid import ObjectId
import linux_cluster as linux

import tests as openstack_driver
from celery import Celery, current_app

app = Flask(__name__)
client = MongoClient("localhost", 27017)
app.config.update(
    CELERY_BROKER_URL="amqp://guest:guest@localhost:5673//",
    CELERY_RESULT_BACKEND="mongodb://localhost:27017/cloud",
)
# Create Celery instance
celery = Celery(
    app.import_name,
    broker=app.config["CELERY_BROKER_URL"],
    backend=app.config["CELERY_RESULT_BACKEND"],
)
celery.conf.update(app.config)
celery.set_default()


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
    if "_id" in template:
        del template["_id"]

    return template


# Endpoint para crear una nueva plantilla / working
@app.route("/slices", methods=["POST"])
def crear_slice():
    new_slice_info = request.json
    new_slice = obtenerTemplateById(new_slice_info["template_id"])
    new_slice.update(new_slice_info)

    db = client.cloud
    collection = db.slices
    result = collection.insert_one(new_slice)
    if result.inserted_id:

        if new_slice["deployment_type"] == "openstack":
            # implementa openstack .

            user_name = request.headers["X-User-Username"]
            openstack_driver.openstackDeployment(new_slice, user_name)

            return jsonify(
                {
                    "msg": f"Slice with id {result.inserted_id} created successfully in OpenStack",
                    "result": "success",
                }
            )

        else:
            # implementa linux
            result_celery = current_app.tasks["linux_cluster.create"].delay(
                str(result.inserted_id)
            )
            return (
                jsonify(
                    {
                        "message": f"Deployment initiated for {result.inserted_id} on Linux Cluster",
                        "task_id": result_celery.id,
                    }
                ),
                202,
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


# Endpoint para buscar una plantilla por ID / working
@app.route("/slices/<string:slice_id>", methods=["GET"])
def buscar_slice(slice_id):
    db = client.cloud
    collection = db.slices
    try:
        slice = collection.find_one({"_id": ObjectId(slice_id)})
        if slice:
            slice_serialized = serialize_document(slice)
            response = jsonify({"result": "success", "slice": slice_serialized})
            return response
        else:
            response = jsonify(
                {
                    "result": "error",
                    "msg": f"Slice with slice id {slice_id} not found",
                }
            )
            error_code = 404
            return response, error_code
    except:
        response = jsonify({"result": "error", "msg": f"Invalid slice id: {slice_id}"})
        error_code = 400
        return response, error_code

# Endpoint para buscar una plantilla por ID / working
@app.route("/slices/vnc/<string:slice_id>", methods=["GET"])
def vnc_slice(slice_id):
    db = client.cloud
    collection = db.slices
    try:
        slice = collection.find_one({"_id": ObjectId(slice_id)})
        if slice:
           
           project_name = slice.get("name")
           vnc_urls = openstack_driver.obtainVNCfromProject(project_name)
           if vnc_urls is None:
                return jsonify({"result": "error", "msg": "Failed to obtain VNC links"}), 500
           
           response = jsonify({"result": "success", "vnc": vnc_urls})
           return response
        else:
            response = jsonify(
                {
                    "result": "error",
                    "msg": f"Slice with slice id {slice_id} not found",
                }
            )
            error_code = 404
            return response, error_code
    except:
        response = jsonify({"result": "error", "msg": f"Invalid slice id: {slice_id}"})
        error_code = 400
        return response, error_code



# Endpoint para eliminar una slice por ID / working
@app.route("/slices/<string:slice_id>", methods=["DELETE"])
def eliminar_slice(slice_id):
    db = client.cloud
    collection = db.slices
    try:
        result = collection.find_one(
            {"_id": ObjectId(slice_id)}, {"deployment_type": 1, "_id": 0, "name":1}
        )
        if result:
            print(result)
            attribute_value = result.get("deployment_type")
            print(f"The value of the attribute deployment_type is: {attribute_value}")
            if attribute_value == "openstack":
                # implementa openstack .
                slice_name_project_name = result.get("name")
                #user_name = request.headers["X-User-Username"]
                deleted_openstack = openstack_driver.openstackDeleteSlice(slice_name_project_name,slice_id)

                if deleted_openstack:

                    result_delete = collection.delete_one({"_id": ObjectId(slice_id)})
                    if result_delete.deleted_count == 1:
                        print(f"Slice with slice id {slice_id} deleted successfully on OpenStack")
                        return jsonify(
                            {
                                "msg": f"Slice with id {slice_id} deleted successfully in OpenStack",
                                "result": "success",
                            }
                    )
                    else:
                        print(f"Slice with slice id {slice_id} not deleted correctly. still in database")
                        return jsonify(
                        {
                            "msg": f"Slice with id {slice_id} not deleted in database. error ocurred",
                            "result": "error",
                        }
                    )
                else:
                    return jsonify(
                        {
                            "msg": f"Slice with id {slice_id} not deleted in database. error ocurred",
                            "result": "error",
                        }
                    )

            else:
                # implementa linux
                result_celery = current_app.tasks["linux_cluster.delete"].delay(
                    slice_id
                )
                return (
                    jsonify(
                        {
                            "message": f"Delete workflow initiated for {slice_id} on Linux Cluster",
                            "task_id": result_celery.id,
                        }
                    ),
                    202,
                )
        else:
            response = jsonify(
                {
                    "result": "error",
                    "msg": f"Slice with slice id {slice_id} not found",
                }
            )
            error_code = 404
            return response, error_code

    except:
        response = jsonify({"result": "error", "msg": f"Error ocurred during elimination slice id: {slice_id}"})
        error_code = 400
        return response, error_code


# Route to get task result
@app.route("/tasks/<task_id>", methods=["GET"])
def get_task_result(task_id):
    task_result = AsyncResult(task_id, app=celery)

    response = {"task_id": task_id, "status": task_result.status}

    if task_result.state == "PENDING":
        response.update({"message": "Task is pending execution."})
    elif task_result.state == "STARTED":
        response.update({"message": "Task has started."})
    elif task_result.state == "SUCCESS":
        response.update(
            {"message": "Task completed successfully.", "result": task_result.result}
        )
    elif task_result.state == "FAILURE":
        response.update(
            {
                "message": "Task failed.",
                "result": str(task_result.result),  # Convert the exception to string
                "traceback": task_result.traceback,  # Include the traceback if available
            }
        )
    elif task_result.state == "RETRY":
        response.update(
            {
                "message": "Task is being retried.",
                "result": str(task_result.result),  # Convert the exception to string
            }
        )

    return jsonify(response), 200


if __name__ == "__main__":
    app.run(debug=True, port=9999)
