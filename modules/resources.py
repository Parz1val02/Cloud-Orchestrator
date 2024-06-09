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
velocidadTX = []
velocidadRX = []


def envioInformacion(info):
    url = "http://10.0.10.2:9898/data"
    headers = {"Content-Type": "application/json"}
    response = requests.post(url, data=info, headers=headers)
    print(response.json())


def findCPU():
    global utilizacionCPU
    command_CPU = "initial=$(cat /proc/stat | grep cpu | awk '{print $5}'); echo $initial; sleep 3; final=$(cat /proc/stat | grep cpu | awk '{print $5}'); echo $final"
    procces_CPU = subprocess.Popen(command_CPU, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    output, error = procces_CPU.communicate()
    output = output.decode('utf-8')
    size = int(len(output.replace("\n"," ").strip(" ").split(" ")))
    informacionCPU1 = output.replace("\n"," ").strip(" ").split(" ")[:int(size/2)]
    informacionCPU2 = output.replace("\n"," ").strip(" ").split(" ")[int(size/2):]
    # CPU 0 y CPU 1
    aux = []
    for i in range(1,int(size/2)):
        delta = int(informacionCPU2[i]) - int(informacionCPU1[i])
        #Tener en cuenta que en 3 s hay 300 jiffys -> esto se podría hacer de manera dinámica
        cpu_value = ((300-delta)/300)*100
        aux.append(cpu_value)
    utilizacionCPU = aux
    print("Informacion del CPU recolectada correctamente")


#Info de la memoria RAM
#"MemoriaUsada(Gb)": , "MemoriaDisponible(Mb)": , "MemoriaTotal(Gb)":
def findMemory() :
    global infoMemoria
    command_Memory = "free -b | awk '/^Mem:/{printf \"%.1f GB , %.1f MB , %.1f GB\\n\", $2/1000000000, $3/1000000, $7/1000000000}'"
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
    numbers = ["9","10","11"]

    for number in numbers:
        try:
            output = findLineStorage(number)
            print(number+" "+output)
            aux = output.replace("\n"," ").replace("   "," ").strip(" ").split(" ")
            if 'G' in aux[0]:
                infoStorage=aux
        except:
            pass


    '''global infoStorage


    try:
        output = findLineStorage(11)
        print(output)
        aux = output.replace("\n"," ").replace("   "," ").strip(" ").split(" ")
        if 'G' in aux[0]:
            infoStorage=aux
    except:
        pass'''


    print("Informacion del almacenamiento recolectada correctamente")


def findLineStorage(lineNumber):
    command_Storage = "lsblk -o FSSIZE,FSUSED,FSUSE% | sed -n '"+lineNumber+"p'"
    process_Storage = subprocess.Popen(command_Storage, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    output, error = process_Storage.communicate()
    output = output.decode('utf-8')
    return output


# "ens3(RX)bps": , "ens3(TX)bps": , "ens4(RX)bps": , "ens4(TX)bps":
def findBandWith():
    global velocidadRX
    global velocidadTX
    #                                   extraer bits recibidos y transmitidos con un margen de 3 segundos
    command_Bandwith = "cat /proc/net/dev | grep -E 'ens3|ens4' | awk '{print $2, $10}'; sleep 3;cat /proc/net/dev | grep -E 'ens3|ens4' | awk '{print $2, $10}'"
    process_Bandwith = subprocess.Popen(command_Bandwith, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    output, error = process_Bandwith.communicate()
    output = output.decode('utf-8')
    size = int(len(output.replace("\n"," ").strip(" ").split(" ")))
    infoRed1 = [int(valor) for valor in output.replace("\n"," ").strip(" ").split(" ")[:int(size/2)]] #pertenecientes a ens3
    infoRed2 = [int(valor) for valor in output.replace("\n"," ").strip(" ").split(" ")[int(size/2):]] #pertenecientes a ens4
    auxTX = []
    auxRX = []
    for i in range(0,int(size/2)):
        delta = infoRed2[i] - infoRed1[i]
        if i%2 == 0:
            auxTX.append(round(float(delta/3.0),1))
        else:
            auxRX.append(round(float(delta/3.0),1))
    velocidadRX = auxRX
    velocidadTX = auxTX
    print("Informacion de red recolectada correctamente")

if __name__ == "__main__":
    while True:
        print("Recolectando informacion")
        #Inicio la busqueda de informacion usando hilos
        hilo_CPU = threading.Thread(target=findCPU)
        hilo_memoria = threading.Thread(target=findMemory)
        hilo_storage = threading.Thread(target=findStorage)
        hilo_TX = threading.Thread(target=findBandWith)
        hilos = [hilo_CPU,hilo_memoria,hilo_storage,hilo_TX]
        for hilo in hilos:
            hilo.start()
        for hilo in hilos:
            hilo.join()
        #Aca debería tener la informacion ya recolectada
        overallInfo = {}
        #CPU
        for i in range(len(utilizacionCPU)):
            overallInfo["Core"+str(i)+"(%)"] = round(utilizacionCPU[i],1)

        core_keys = [key for key in overallInfo.keys() if key.startswith("Core")]
        total_cpu_usage = sum(overallInfo[key] for key in core_keys)
        average_cpu_usage = total_cpu_usage / len(core_keys)

        overallInfo["average_cpu_usage (%)"] = average_cpu_usage

        #Memoria
        overallInfo["MemoriaUsada(Gb)"]= float(infoMemoria[6]) if float(infoMemoria[3])>float(infoMemoria[6]) else   float(infoMemoria[3])
        overallInfo["MemoriaDisponible(Mb)"]= float(infoMemoria[3]) if float(infoMemoria[3])>float(infoMemoria[6]) else   float(infoMemoria[6])
        overallInfo["MemoriaTotal(Gb)"]=float(infoMemoria[0])
        #Almacenamiento
        print(infoStorage)
        overallInfo["AlmacenamientoUsado(Gb)"]=float(infoStorage[1].strip("G"))
        overallInfo["AlmacenamientoUsado(%)"]=int(infoStorage[3].strip("%"))
        overallInfo["AlmacenamientoTotal(Gb)"]=float(infoStorage[0].strip("G"))
        #Red
        overallInfo["ens3(RX)bps"]=velocidadRX[0]
        overallInfo["ens3(TX)bps"]=velocidadTX[0]
        overallInfo["ens4(RX)bps"]=velocidadRX[1]
        overallInfo["ens4(TX)bps"]=velocidadTX[1]
        #Timestamp
        overallInfo["timestamp"]=datetime.now().strftime("%d-%m-%Y %H:%M:%S")

        print("Informacion recolectada correctamente")
        print("Enviando la información al servidor")
        #envioInformacion(json.dumps(overallInfo))
        print("Informacion enviada correctamente!")
        #print(overallInfo)
        print(json.dumps(overallInfo))
        envioInformacion(json.dumps(overallInfo))
        time.sleep(5)