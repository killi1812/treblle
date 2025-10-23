import serverApi from "./serverAxios";

/**
 * Exchanges a Discord authorization code for an access token by making a request to the server's /token endpoint.
 * This is a key step in the OAuth2 flow, allowing the client to obtain a token to authenticate with the Discord API on behalf of the user.
 *
 * @param {string} code - The authorization code received from the Discord SDK's authorize command.
 * @returns {Promise<string | undefined>} A promise that resolves to the access token if successful, otherwise undefined.
 */
export async function authorizeServer(code: string): Promise<string | undefined> {
  try {
    const rez = await serverApi.post<{ access_token: string }>("/token", {
      code: code,
    })
    console.log(rez.data)

    return rez.data.access_token
  }
  catch (error: any) {
    console.error(error)
  }
}
