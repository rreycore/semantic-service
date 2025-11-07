import asyncio
import logging

import pgai

from .cfg import cfg


async def main():
    await pgai.ainstall(cfg.get_db_url())


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    asyncio.run(main())
