import os

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DATA_DIR = os.path.join(BASE_DIR, 'data')
ARTIFACTS_DIR = os.path.join(BASE_DIR, 'artifacts')

EPOCHS = 50
BATCH_SIZE = 32
VALIDATION_SPLIT = 0.1

# All datasets should have the same structure after mapping, with columns: timestamp, cpu_usage, ram_usage
DATASETS_CONFIG = [
    {
        "id": "azure_v2_vm_cpu_readings_month_aggregated_cpu_mem",
        "url": "https://github.com/alejandrofdez-us/DataCenter-Traces-Datasets/raw/refs/heads/main/azure_v2/vm_cpu_readings_month_aggregated_cpu_mem.csv",
    	"mapping": {
            "timestamp": "timestamp",
            "cpu_usage": "cpu_usage",
            "assigned_mem": "ram_usage"
        },
        "time_unit": "s"
    },
    {
        "id": "azure_v2_vm_cpu_readings_week1_aggregated_cpu_mem",
        "url": "https://github.com/alejandrofdez-us/DataCenter-Traces-Datasets/raw/refs/heads/main/azure_v2/week1/vm_cpu_readings_week1_aggregated_cpu_mem.csv",
        "mapping": {
            "timestamp": "timestamp",
            "cpu_usage": "cpu_usage",
            "assigned_mem": "ram_usage"
        },
        "time_unit": "s"
    },
    {
        "id": "azure_v2_vm_cpu_readings_week2_aggregated_cpu_mem",
        "url": "https://github.com/alejandrofdez-us/DataCenter-Traces-Datasets/raw/refs/heads/main/azure_v2/week2/vm_cpu_readings_week2_aggregated_cpu_mem.csv",
        "mapping": {
            "timestamp": "timestamp",
            "cpu_usage": "cpu_usage",
            "assigned_mem": "ram_usage"
        },
        "time_unit": "s"
    },
    {
        "id": "azure_v2_vm_cpu_readings_week3_aggregated_cpu_mem",
        "url": "https://github.com/alejandrofdez-us/DataCenter-Traces-Datasets/raw/refs/heads/main/azure_v2/week3/vm_cpu_readings_week3_aggregated_cpu_mem.csv",
        "mapping": {
            "timestamp": "timestamp",
            "cpu_usage": "cpu_usage",
            "assigned_mem": "ram_usage"
        },
        "time_unit": "s"
    },
    {
        "id": "azure_v2_vm_cpu_readings_week4_aggregated_cpu_mem",
        "url": "https://github.com/alejandrofdez-us/DataCenter-Traces-Datasets/raw/refs/heads/main/azure_v2/week4/vm_cpu_readings_week4_aggregated_cpu_mem.csv",
        "mapping": {
            "timestamp": "timestamp",
            "cpu_usage": "cpu_usage",
            "assigned_mem": "ram_usage"
        },
        "time_unit": "s"
    },
    {
        "id": "azure_v2_vm_cpu_readings_week5_aggregated_cpu_mem",
        "url": "https://github.com/alejandrofdez-us/DataCenter-Traces-Datasets/raw/refs/heads/main/azure_v2/week5/vm_cpu_readings_week5_aggregated_cpu_mem.csv",
        "mapping": {
            "timestamp": "timestamp",
            "cpu_usage": "cpu_usage",
            "assigned_mem": "ram_usage"
        },
        "time_unit": "s"
    },
#     {
#         "id": "my_macbook_docker_cpu_mem_2026-03-17",
#         "url": "my_macbook_docker_cpu_mem_2026-03-17.csv",
#         "mapping": {
#             "timestamp": "timestamp",
#             "cpu_usage": "cpu_usage",
#             "assigned_mem": "ram_usage"
#         },
#         "time_unit": "s"
#     },
#     {
#         "id": "my_macbook_docker_cpu_mem_test-node-01_2026-03-30",
#         "url": "my_macbook_docker_cpu_mem_test-node-01_2026-03-30.csv",
#         "mapping": {
#             "timestamp": "timestamp",
#             "cpu_usage": "cpu_usage",
#             "assigned_mem": "ram_usage"
#         },
#         "time_unit": "s"
#     },
#     {
#         "id": "my_macbook_docker_cpu_mem_test-node-02_2026-03-30",
#         "url": "my_macbook_docker_cpu_mem_test-node-02_2026-03-30.csv",
#         "mapping": {
#             "timestamp": "timestamp",
#             "cpu_usage": "cpu_usage",
#             "assigned_mem": "ram_usage"
#         },
#         "time_unit": "s"
#     }
]
