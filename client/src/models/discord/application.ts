export interface Application {
  id: string
  name: string
  icon?: string | null | undefined
  rpc_origins?: string[] | undefined
}
