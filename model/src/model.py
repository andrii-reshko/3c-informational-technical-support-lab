import tensorflow as tf
from tensorflow.keras import layers, models


# Pinball loss (aka quantile loss) is used for training quantile regression models.
# It penalizes underestimation and overestimation differently based on the specified quantiles.
# In this case, we are calculating the loss for the 10% and 90% quantiles,
# which correspond to the lower and upper bounds of our prediction interval, respectively.
def pinball_loss(y_true, y_pred):
    q_low, q_high = 0.1, 0.9

    err_low = y_true - y_pred[:, 0:1]
    err_high = y_true - y_pred[:, 1:2]

    loss_low = tf.reduce_mean(tf.maximum(q_low * err_low, (q_low - 1) * err_low))
    loss_high = tf.reduce_mean(tf.maximum(q_high * err_high, (q_high - 1) * err_high))

    return loss_low + loss_high


def build_mlp(input_dim: int) -> models.Model:
    model = models.Sequential([
        layers.Input(shape=(input_dim,)),
        layers.Dense(64, activation='relu'),
        layers.Dense(32, activation='relu'),
        layers.Dense(16, activation='relu'),
        layers.Dense(2)  # upper and lower bounds (P10 and P90)
    ])

    model.compile(optimizer='adam', loss=pinball_loss)
    return model
