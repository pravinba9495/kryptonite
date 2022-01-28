import Axios from "axios";

/**
 * Kraken API
 */
export class Kraken {
  /**
   * GetCoinPrice returns the coin price in USD
   * @param pair Ticker symbol in Kraken
   * @returns Promise<AxiosResponse<any, any>>
   */
  static GetCoinPrice = async (pair: string): Promise<number> => {
    return Axios.get<number>("https://api.kraken.com/0/public/Ticker", {
      params: {
        pair,
      },
    })
      .then((response) => response.data)
      .then((response: any) => {
        if (response.error.length > 0) {
          return Promise.reject(response.error[0]);
        }
        return Number(response["result"][pair.toUpperCase()]["a"][0]);
      })
      .catch((error) => {
        if (error.response) {
          return Promise.reject(error.response.data);
        } else {
          return Promise.reject(error.request);
        }
      });
  };
}
