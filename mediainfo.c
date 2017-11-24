#include <libavutil/avutil.h>
#include <libavutil/avstring.h>
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavformat/avio.h>

typedef struct buffer_data {
    uint8_t *ptr;
    size_t size; ///< size left in the buffer
    char*  name;
} buffer_data;

buffer_data* alloc_buffer_data(uint8_t *buf, size_t buf_size) {
  buffer_data* bd;
  bd = malloc(sizeof(buffer_data));
  bd->ptr = buf;
  bd->size = buf_size;
  bd->name = "bla";

  return bd;
}

void free_buffer_data(buffer_data *bd){
  free(bd);
}

int read_packet(void *opaque, uint8_t *buf, int buf_size)
{
    struct buffer_data *bd = (struct buffer_data *)opaque;
    buf_size = FFMIN(buf_size, bd->size);
    printf("ptr:%p size:%zu\n", bd->ptr, bd->size);
    /* copy internal buffer data to buf */
    memcpy(buf, bd->ptr, buf_size);
    bd->ptr  += buf_size;
    bd->size -= buf_size;
    return buf_size;
}

