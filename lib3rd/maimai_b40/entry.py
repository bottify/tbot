import maimai_best_40
import io, sys, asyncio

import math
from PIL import Image
# payload
# {qq: xxx}
# {username: xxx}

async def main():
    # img = Image.new('RGB', (10, 10), (0, 0, 0))
    # for x in range(10):
    #     for y in range(10):
    #         c = math.floor(x / 10 * 255)
    #         img.putpixel((x, y), (c,c,c))
    img, ret = await maimai_best_40.generate({sys.argv[1]: sys.argv[2]})
    if ret != 0:
        return
    bo = io.BytesIO()
    img.save(bo, 'png')
    sys.stdout.buffer.write(bo.getvalue())

if len(sys.argv) < 3:
    exit()
asyncio.run(main())
