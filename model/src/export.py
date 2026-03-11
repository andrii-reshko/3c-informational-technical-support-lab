import tensorflow as tf
import json
import os


def export_pipeline(keras_model, scaler, feature_names, node_id, resource, horizon, output_dir, quality_metrics):
    os.makedirs(output_dir, exist_ok=True)

    # Common name like: node-01_cpu_usage_15m_model.tflite
    base_name = f"{node_id}_{resource}_{horizon}m"

    # Export to TFLite
    tflite_path = os.path.join(output_dir, f"{base_name}_model.tflite")
    converter = tf.lite.TFLiteConverter.from_keras_model(keras_model)
    with open(tflite_path, 'wb') as f:
        f.write(converter.convert())

    # Export scaler parameters to JSON (we'll need these for normalization in Go)
    scaler_data = {
        "features": feature_names.tolist(),
        "scale": scaler.scale_.tolist(),
        "min": scaler.min_.tolist(),
        "data_min": scaler.data_min_.tolist(),
        "data_max": scaler.data_max_.tolist()
    }
    scaler_path = os.path.join(output_dir, f"{base_name}_scaler.json")
    with open(scaler_path, 'w') as f:
        json.dump(scaler_data, f, indent=4)

    # Export meta data about model quality and configuration
    meta_data = {
        "mae": quality_metrics["mae"],
        "coverage": quality_metrics["coverage"],
        "resource": resource,
        "horizon": horizon,
        "features_count": len(feature_names)
    }
    meta_path = os.path.join(output_dir, f"{base_name}_meta.json")
    with open(meta_path, 'w') as f:
        json.dump(meta_data, f, indent=4)

    print(f"exported:")
    print(f" - {tflite_path}")
    print(f" - {tflite_path}")
    print(f" - {tflite_path}")

