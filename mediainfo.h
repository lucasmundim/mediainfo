#include <libavutil/avutil.h>
#include <libavutil/avstring.h>
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavformat/avio.h>

typedef struct buffer_data {
    uint8_t *ptr;
    size_t size; ///< size left in the buffer
} buffer_data;

int read_packet(void *opaque, uint8_t *buf, int buf_size);
buffer_data* alloc_buffer_data(uint8_t *buf, size_t buf_size);
void free_buffer_data(buffer_data *bd);
