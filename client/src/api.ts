import axios from "axios";

interface PrepareResponse {
  id: string;
}

const api = {
  Prepare: async (name: string, contentType: string) => {
    try {
      return await axios.post<PrepareResponse>("/file/prepare", {
        name,
        content_type: contentType,
      });
    } catch (error) {
      console.error(error);
    }
  },
};

export default api;
