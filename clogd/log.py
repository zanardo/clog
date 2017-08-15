# -*- coding: utf-8 -*-
#
# Copyright (c) 2013-2017, Antonio Zanardo <zanardo@gmail.com>
#

import logging

logging.basicConfig(
    level=logging.DEBUG, format='%(asctime)s.%(msecs)03d '
    '%(levelname)3s | %(message)s', datefmt='%Y/%m/%d %H:%M:%S'
)
log = logging.getLogger(__name__)
