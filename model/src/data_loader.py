import pandas as pd
import os
import urllib.request
from src.config import DATASETS_CONFIG, DATA_DIR


def load_trace_data() -> pd.DataFrame:
    os.makedirs(DATA_DIR, exist_ok=True)
    all_dataframes = []

    for config in DATASETS_CONFIG:
        dataset_id = config["id"]
        url = config.get("url", "")
        file_ext = config.get("format", "csv")

        file_path = os.path.join(DATA_DIR, f"{dataset_id}.{file_ext}")

        if not os.path.exists(file_path):
            if url:
                print(f"  Loading {dataset_id} from URL {url}")
                try:
                    urllib.request.urlretrieve(url, file_path)
                except Exception as e:
                    print(f"  Error loading {dataset_id}: {e}")
                    continue
            else:
                print(f"  File not found {dataset_id}, download it manually and place in {DATA_DIR}/")
                continue
        else:
            print(f"  Local dataset {dataset_id} found at {file_path}")

        print(f"  Reading dataset {dataset_id} as {file_ext}")
        try:
            if file_ext == "csv":
                df = pd.read_csv(file_path)
            else:
                raise ValueError(f"  Unsupported file format: {file_ext}")

        except Exception as e:
            print(f"  Reading error {dataset_id}: {e}")
            continue

        df = df.rename(columns=config["mapping"])
        expected_cols = list(config["mapping"].values())

        missing_cols = [col for col in expected_cols if col not in df.columns]
        if missing_cols:
            print(f"  Missing columns {missing_cols} in dataset {dataset_id}, available {list(df.columns)}")
            print(f"  - skipping")
            continue

        df = df[expected_cols].copy()
        df['timestamp'] = pd.to_datetime(df['timestamp'], unit=config.get('time_unit', 's'), origin='2023-01-01')

        for col in expected_cols:
            if col in df.columns and col != 'timestamp':
                max_val = df[col].max()
                if max_val > 0:
                    df[col] = (df[col] / max_val) * 100.0

        df['trace_id'] = dataset_id
        all_dataframes.append(df)

    if not all_dataframes:
        raise ValueError("could not load any dataset")

    combined_df = pd.concat(all_dataframes, ignore_index=True)
    combined_df.set_index('timestamp', inplace=True)

    print(f"  Processed {len(combined_df)} records")
    return combined_df
