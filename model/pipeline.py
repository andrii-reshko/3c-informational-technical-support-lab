import numpy as np
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import MinMaxScaler
from src.config import ARTIFACTS_DIR, EPOCHS, BATCH_SIZE, VALIDATION_SPLIT
from src.data_loader import load_trace_data
from src.features import build_features
from src.model import build_mlp
from src.export import export_pipeline


def train_single_model(df, node_id, resource, horizon):
    print(f"training: {resource} for horizon {horizon} min\n")

    dataset = build_features(df, target_col=resource, horizon=horizon)

    x = dataset.drop(columns=['target_peak'])
    y = dataset[['target_peak']]
    feature_names = x.columns

    scaler = MinMaxScaler()
    x_scaled = scaler.fit_transform(x)

    x_train, x_test, y_train, y_test = train_test_split(x_scaled, y, test_size=0.2, shuffle=False)

    model = build_mlp(input_dim=x_train.shape[1])

    model.fit(
        x_train, y_train,
        epochs=EPOCHS,
        batch_size=BATCH_SIZE,
        validation_split=VALIDATION_SPLIT,
        verbose=0
    )

    # Get predictions on the test split
    y_predictions = model.predict(x_test)
    y_true = y_test.values

    # Compute MAE (for peak forecasting we usually use the upper bound or mean).
    # In the specification, MAE is calculated between actual and predicted values
    # use upper bound (peak)
    mae = np.mean(np.abs(y_true - y_predictions[:, 1:2]))

    # Compute coverage (percentage of actual values inside [q0.1, q0.9]).
    # Key trust metric in the specification
    within_interval = (y_true >= y_predictions[:, 0:1]) & (y_true <= y_predictions[:, 1:2])
    coverage = np.mean(within_interval)
    loss = model.evaluate(x_test, y_test, verbose=0)
    print(f"  pinball loss: {loss:.4f}")
    print(f"  mae: {mae:.4f}")
    print(f"  coverage: {coverage:.4f}")

    quality_metrics = {
        "mae": float(mae),
        "coverage": float(coverage),
        "loss": float(loss)
    }

    export_pipeline(model, scaler, feature_names, node_id, resource, horizon, ARTIFACTS_DIR, quality_metrics)


def run_all():
    print("\nstart pipeline")
    node_id = "global_trace"

    # Call without arguments; the loader already knows where config is
    df = load_trace_data()

    # Config model(s) characteristics
    target_resources = ['cpu_usage', 'ram_usage']
    horizons_minutes = [15]

    for resource in target_resources:
        for horizon in horizons_minutes:
            train_single_model(df, node_id, resource, horizon)

    print("\nall done")


if __name__ == "__main__":
    run_all()
