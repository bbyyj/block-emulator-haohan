import argparse
import os
import time

import paramiko
from scp import SCPClient
from threading import Thread

# ExpRequiredSignNum = 2
# ExpProposerNum = 2
# ExpTotalSignerNum = 3
# ExpNodesNum = 16
# ExpPreloadDataSize = 3000000
# ExpBlockSize = 10000
# ExpRunTime = 200
# ports = {50002,50004,50005,50006,50007,50008,50009,50010}

# def set_ports():
#     global ports
#     machineNumber = min(ExpNodesNum, 13)
#     ports = list()
#     for i in range(machineNumber):
#         port = 50001+i
#         ports.append(port)

ports = {50001, 50002, 50003, 50005, 50006, 50007, 50008, 50009, 50010}
def scp_files_to_remote():
    i = 1
    for port in ports:
        ssh = paramiko.SSHClient()
        ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())  # 自动添加主机名及主机密钥到本地HostKeys对象，并保存，只在代理模式下使用
        print('Connecting=======')
        ssh.connect('172.16.108.106', username='huanglab', password='huanglab', port=port)  # 也可以使用key_filename参数提供私钥文件路径
        print('Connected!!!!!!!!')
        with SCPClient(ssh.get_transport()) as scp:
            scp.put('./haohanBin', recursive=True, remote_path='~/haohanWorkSpace/block-emulator-main/')
            # scp.put('./bEmulator_LW/ip.json', recursive=True, remote_path=' ~/haohanWorkSpace/block-emulator-main/')
            scp.close()
        ssh.close()
        i += 1


def download_from_remote(remote_path, local_path):
    # Check if local path exists, if not, create it
    if not os.path.exists(local_path):
        os.makedirs(local_path)
        print(f'Created local path: {local_path}')

    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    print('Connecting=======')
    ssh.connect('172.16.108.106', username='huanglab', password='huanglab',
                port=50001)  # You can also use the key_filename parameter to provide the path to the private key file
    print('Connected!!!!!!!!')

    with SCPClient(ssh.get_transport()) as scp:
        scp.get(remote_path, local_path, recursive=True)
        scp.close()

    ssh.close()



def run_exe_remote(port, cmd,master):
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())  # 自动添加主机名及主机密钥到本地HostKeys对象，并保存，只在代理模式下使用
    print('Connecting=======')
    ssh.connect('172.16.108.106', username='huanglab', password='huanglab', port=port)  # 也可以使用key_filename参数提供私钥文件路径
    stdin, stdout, stderr = ssh.exec_command(cmd)
        
    # 输出命令执行结果
    print(f"Standard output for command '{cmd}':")
    if master == 1:
        for line in stdout:
            print(line.strip())

    
    # 输出错误信息
    print(f"Standard error for command '{cmd}':"+" "+str(port))
    for line in stderr:
        print(line.strip())
    
    ssh.close()


def multi_Clients_runing(shardNum,NodeNum,algorithm,totalDataSize,injectSpeed,maxBlockSize_global,block_Interval):
    threads = []  # 用于存储所有线程对象
    # 先运行打包交易的节点
    base_port = 50002
    base_cmd = "cd ~/haohanWorkSpace/block-emulator-main/ && sh bat_shardNum={}_NodeNum={}_mod=Relay_{}.sh {} {} {} {} {}"

    for i in range(8):
        port = base_port + i
        cmd = base_cmd.format(shardNum,NodeNum,port,algorithm,totalDataSize,injectSpeed,maxBlockSize_global,block_Interval)
        newPort=port
        if (port == 50004):
            newPort = 50010
        client = Thread(target=run_exe_remote, args=(newPort, cmd,0))
        threads.append(client)  # 将线程对象添加到列表中
        client.start()

    # 运行supervisor
    time.sleep(10)
    super_port = 50001
    super_cmd = "cd ~/haohanWorkSpace/block-emulator-main/ && ./haohanBin -c -S {} -N {} -m 3 -a {} -d {} -i {} -b {} -t {} "
    super_cmd = super_cmd.format(shardNum,NodeNum,algorithm,totalDataSize,injectSpeed,maxBlockSize_global,block_Interval)
    super_client = Thread(target=run_exe_remote, args=(super_port, super_cmd,1))
    threads.append(super_client)  # 将线程对象添加到列表中
    super_client.start()

    # 等待所有线程执行完毕
    for thread in threads:
        thread.join()





def clear_history_expData():
    ports = {50001, 50002, 50003, 50005, 50006, 50007, 50008, 50009, 50010}
    for port in ports:
        run_exe_remote(port=port, cmd=" rm -rf ~/haohanWorkSpace/block-emulator-main/expTest",master=0)
        run_exe_remote(port=port, cmd=" killall -9 haohanBin",master=0)
        run_exe_remote(port=port,
                       cmd=" cd ~/haohanWorkSpace/block-emulator-main/ && fuser -k haohanBin",master=0)
        # run_exe_remote(port=port, cmd=" ps -ef| grep build | cut -c 9-16|xargs kill -9")
        


# def mkdir_exp_dir():
#     for port in ports:
#         run_exe_remote(port=port, cmd="mkdir -p  ~/haohanWorkSpace/block-emulator-main/")
#
#
#
# def chmod_executionFile():
#     for port in ports:
#         run_exe_remote(port=port, cmd="cd  ~/haohanWorkSpace/block-emulator-main/ && chmod +x ./bEmulatorFeather")



# -----------------------------------------------------------------------------
# loop for experiments

# 1. 实验1：改变交易的数量
# default config
shardNum = [8]
nodeNum = 4
algorithm =[2]
totalDataSize =[900000]#[800000 , 900000 , 1000000]
injectSpeed = 5000
maxBlockSize_global =2000
block_Interval =6000

for shard in shardNum:
    for datasize in totalDataSize:
        for algo in algorithm:
            print("--------------------------------------------------------------------------------------------------------")
            print(
                "--------------------------------------------------------------------------------------------------------")
            clear_history_expData()
            # scp_files_to_remote()
            multi_Clients_runing(shard,nodeNum,algo,datasize,injectSpeed,maxBlockSize_global,block_Interval)
            # local_path = "./Emulator/tx_number/t2_shardNum={}_nodeNum={}_algorithm={}_totalDataSize={}_injectSpeed={}_maxBlockSize_global={}_block_Interval={}".format(
            #     shard, nodeNum, algo, datasize, injectSpeed, maxBlockSize_global, block_Interval
            # )
            # download_from_remote("~/haohanWorkSpace/block-emulator-main/result",local_path)



