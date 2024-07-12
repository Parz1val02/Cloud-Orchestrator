
import socket
import time
import subprocess
import threading
import json
from datetime import datetime
import requests


utilizacionCPU = []
infoMemoria = []
infoStorage = []


def envioInformacion(info):
    url = "http://10.0.10.2:9898/data"
    headers = {"Content-Type": "application/json"}
    response = requests.post(url, data=info, headers=headers)
    print(response.json())


def findCPU():
    global utilizacionCPU
    command_CPU = "initial=$(cat /proc/stat | grep cpu | awk '{print $5}'); echo $initial; sleep 0.5; final=$(cat /proc/stat | grep cpu | awk '{print $5}'); echo $final"
    procces_CPU = subprocess.Popen(command_CPU, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    output, error = procces_CPU.communicate()
    output = output.decode('utf-8')
    lsIDLE =  output.replace("\n"," ").strip(" ").split(" ")
    lsIDLE.pop(0)
    lsIDLE.pop(8)
    size = len(lsIDLE)
    mitad = size//2
    parte1 = lsIDLE[:mitad]
    parte2 = lsIDLE[mitad:]
    print(parte1)
    print(parte2)
    #informacionCPU1 = [lsIDLE[1],lsIDLE[10]]
    aux = []
    for i in range(size//2):
        #encontramos los porcetnajes de CPU para los 0.5 seg examinados
        delta=int(parte2[i]) - int(parte1[i])
        diff=50-delta
        if diff < 0:
            diff=0
        cpu_used = (diff/50)*100
        aux.append(cpu_used)
    utilizacionCPU = aux
    print("Informacion de CPU recolectada correctamente")

#Info de la memoria RAM
#"MemoriaUsada(Gb)": , "MemoriaDisponible(Mb)": , "MemoriaTotal(Gb)":
def findMemory() :
    global infoMemoria
    command_Memory = "free -b | awk '/^Mem:/{printf \"%.1f MB , %.1f MB , %.1f MB\\n\", $2/1000000, $3/1000000, $7/1000000}'"
    process_Memory = subprocess.Popen(command_Memory, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    output, error = process_Memory.communicate()
    output = output.decode('utf-8')
    infoMemoria = output.replace("\n"," ").strip(" ").split(" ")
    print("Informacion de la memoria recolectada correctamente")
    #Recordar 1° -> Memoria Total, 2° -> Memoria Usada, 3° -> Memoria Disponible


# "MemoriaUsada(Gb)": , "MemoriaDisponible(Mb)": , "MemoriaTotal(Gb)":
# "AlmacenamientoUsado(Gb)": , "AlmacenamientoUsado(%)": , "AlmacenamientoTotal(Gb)":
def findStorage():
    global infoStorage
    try:
       output = findLineStorage("11")
       aux = output.replace("\n"," ").replace("   "," ").strip(" ").split(" ")
       if 'G' in aux[0]:
           infoStorage = aux
       else:
           infostorage = []
    except Exception as e:
       print(f"Error processing line 11: {e}")
       infoStorage = []



    print("Informacion del almacenamiento recolectada correctamente")
    print(infoStorage)
	
def findLineStorage(lineNumber):
    command_Storage = "lsblk -o FSSIZE,FSUSED,FSUSE% | sed -n '"+lineNumber+"p'"
    process_Storage = subprocess.Popen(command_Storage, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    output, error = process_Storage.communicate()
    output = output.decode('utf-8')
    return output



if __name__ == "__main__":
    while True:
        print("Recolectando informacion")
        #Inicio la busqueda de informacion usando hilos
        hilo_CPU = threading.Thread(target=findCPU)
        hilo_memoria = threading.Thread(target=findMemory)
        hilo_storage = threading.Thread(target=findStorage)
        #hilo_TX = threading.Thread(target=findBandWith)
        hilos = [hilo_CPU,hilo_memoria,hilo_storage]
        for hilo in hilos:
            hilo.start()
        for hilo in hilos:
            hilo.join()
        #Ordenamos la data recopilada para enviarla
        overallInfo = {}
        overallInfo["worker1"] =  "10.0.0.30"
        #CPU
        for i in range(len(utilizacionCPU)):
            overallInfo["Core"+str(i)+"(%)"] = round(utilizacionCPU[i],1)

        #Memoria
        #print(infoMemoria)
        overallInfo["MemoriaUsada(Mb)"]= float(infoMemoria[3])
        overallInfo["MemoriaDisponible(Mb)"]= float(infoMemoria[6])
        overallInfo["MemoriaTotal(Mb)"]=float(infoMemoria[0])
        #Almacenamiento
        #print(infoStorage)
        overallInfo["AlmacenamientoUsado(Gb)"]=float(infoStorage[1].strip("G"))
        overallInfo["AlmacenamientoUsado(%)"]=int(infoStorage[3].strip("%"))
        overallInfo["AlmacenamientoTotal(Gb)"]=float(infoStorage[0].strip("G"))
        #Timestamp
        overallInfo["timestamp"]=datetime.now().strftime("%d-%m-%Y %H:%M:%S")

        print("Informacion recolectada correctamente")
        print("Enviando la información al servidor")
        #print(overallInfo)
        print(json.dumps(overallInfo))
        envioInformacion(json.dumps(overallInfo))
        time.sleep(0.5)
