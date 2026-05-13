# Схема зв'язків між класами та компонентами системи

```mermaid
classDiagram
    class AppContainer {
        +DB sqlite.DB
        +PrometheusClient prometheus.Client
        +NodeRepo NodeRepository
        +MetricsRepo MetricsRepository
        +Forecaster ForecasterService
        +Collector CollectorService
    }
    
    class ForecasterService {
        +GetForecast(nodeID, horizon) ForecastResponse
        +cpuModel TFLiteModel
        +ramModel TFLiteModel
        +predictResource() resourcePrediction
    }
    
    class resourcePrediction {
        +current float64
        +lower float64
        +upper float64
        +predicted float64
        +quality ModelQuality
    }
    
    class FeatureEngineer {
        +PrepareVector(metrics, isRAM) FeatureVector
    }
    
    class TFLiteModel {
        +Predict(features) [lower, upper]
    }
    
    class CollectorService {
        +CollectMetrics()
        +SyncNodes()
    }
    
    class RestAPI {
        +PredictHandler()
        +UIHandler()
        +WebSocketHandler()
    }

    AppContainer --> ForecasterService
    AppContainer --> CollectorService
    AppContainer --> NodeRepo
    AppContainer --> MetricsRepo

    ForecasterService --> FeatureEngineer
    ForecasterService --> TFLiteModel
    ForecasterService --> MetricsRepo
    ForecasterService --> resourcePrediction

    CollectorService --> MetricsRepo
    RestAPI --> ForecasterService
    RestAPI --> MetricsRepo
```
