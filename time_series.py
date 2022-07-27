import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import statsmodels.api as sm
from statsmodels.tsa.api import ExponentialSmoothing, Holt
from sklearn.metrics import mean_squared_error
from math import sqrt

df = pd.read_csv('/Users/dimabyckov/Desktop/Работа/DataSets/Holt-ARIMA', sep=",")
df.head()
df.tail()
df = pd.read_csv('/Users/dimabyckov/Desktop/Работа/DataSets/Holt-ARIMA', sep=",")
Start_Train = int(input())
End_Train = int(input())
train = df[Start_Train:End_Train]
test = df[End_Train:]



train.Count.plot(figsize=(15, 8), title='Daily Ridership', fontsize=14)
test.Count.plot(figsize=(15, 8), title='Daily Ridership', fontsize=14)
plt.show()

sm.tsa.seasonal_decompose(train.Count).plot()
result = sm.tsa.stattools.adfuller(train.Count)
plt.show()

y_hat_avg = test.copy()
fit1 = Holt(np.asarray(train['Count'])).fit(smoothing_level=0.3, smoothing_slope=0.1)
y_hat_avg['Holt_linear'] = fit1.forecast(len(test))
plt.figure(figsize=(16, 8))
plt.plot(train['Count'], label='Тренировка')
plt.plot(test['Count'], label='Тест')
plt.plot(y_hat_avg['Holt_linear'], label='Линейный тренд Холта')
plt.legend(loc='best')
plt.title("Линейный тренд Холта")
plt.show()


y_hat_avg = test.copy()
fit1 = ExponentialSmoothing(np.asarray(train['Count']), seasonal_periods=7, trend='add', seasonal='add', ).fit()
y_hat_avg['Holt_Winter'] = fit1.forecast(len(test))
plt.figure(figsize=(16, 8))
plt.plot(train['Count'], label='Тренировка')
plt.plot(test['Count'], label='Тест')
plt.plot(y_hat_avg['Holt_Winter'], label='Метод Холта-Винтера')
plt.legend(loc='best')
plt.title("Метод Холта-Винтера")
plt.show()

rms = sqrt(mean_squared_error(test.Count, y_hat_avg.Holt_Winter))
print(rms)


y_hat_avg = test.copy()
fit1 = sm.tsa.statespace.SARIMAX(train.Count, order=(7, 1, 4), seasonal_order=(0, 1, 2, 7)).fit()
y_hat_avg['SARIMA'] = fit1.predict(start="2013-11-1", end="2013-12-31", damped=True)
plt.figure(figsize=(16, 8))
plt.plot(train['Count'], label='Тренировка')
plt.plot(test['Count'], label='Тест')
plt.plot(y_hat_avg['SARIMA'], label='SARIMA')
plt.legend(loc='best')
plt.title("SARIMA")
plt.show()

rms = sqrt(mean_squared_error(test.Count, y_hat_avg.SARIMA))
print(rms)
