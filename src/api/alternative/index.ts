import Axios from "axios";

export class Alternative {
  static GetCryptoFearIndex = async (): Promise<any> => {
    return Axios.get<number>("https://api.alternative.me/fng/", {
      timeout: 5000,
    })
      .then((response) => response.data)
      .then((response: any) => {
        const { value, value_classification } = response["data"][0];
        return {
          fearGreedIndex: Number(value),
          fearGreedIndexClassification: value_classification,
        };
      })
      .catch((error) => {
        if (error.response && error.response.data) {
          return Promise.reject(error.response.data);
        } else {
          return Promise.reject(new Error("Request failed"));
        }
      });
  };
}
