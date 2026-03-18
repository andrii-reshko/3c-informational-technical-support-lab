import pandas as pd


def build_features(df: pd.DataFrame, target_col: str, horizon: int) -> pd.DataFrame:
    data = df[[target_col, 'trace_id']].copy()

    lags = [1, 2, 3, 5, 10, 15, 30]
    if horizon not in lags:
        lags.append(horizon)
    lags = sorted(list(set(lags)))

    grouped = data.groupby('trace_id')[target_col]

    # Lags
    for lag in lags:
        data[f'lag_{lag}'] = grouped.shift(lag)

    # Multiple rolling windows
    windows = [5, 15, 60]
    for w in windows:
        data[f'rolling_mean_{w}'] = grouped.transform(lambda x: x.rolling(window=w).mean())
        data[f'rolling_max_{w}'] = grouped.transform(lambda x: x.rolling(window=w).max())

    # Time-based features
    data['hour'] = data.index.hour
    data['day_of_week'] = data.index.dayofweek

    # Target: max(resource) in next 'horizon' minutes
    data['target_peak'] = grouped.transform(lambda x: x.rolling(window=horizon).max().shift(-horizon))

    data = data.drop(columns=['trace_id'])
    return data.dropna()
