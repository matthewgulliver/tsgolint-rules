// expect: no-outside-declaration-in-the-hexagon
import { PostgresPledgeRepository } from "../../adapters/driven/postgres/src/repo"
export const repository = new PostgresPledgeRepository()
