# Simple usage with a mounted data directory:
# > docker build -t gaia .
# > docker run -v $HOME/.gaiad:/root/.gaiad gaia init
# > docker run -v $HOME/.gaiad:/root/.gaiad gaia start

FROM alpine:edge

# Set up dependencies
ENV PACKAGES go glide make git libc-dev bash

# Set up GOPATH & PATH
ENV GOPATH       /root/go
ENV BASE_PATH    $GOPATH/src/github.com/cosmos
ENV REPO_PATH    $BASE_PATH/cosmos-sdk
ENV WORKDIR      /cosmos/
ENV PATH         $GOPATH/bin:$PATH

# Link expected Go repo path
RUN mkdir -p $WORKDIR $GOPATH/pkg $ $GOPATH/bin $BASE_PATH

#Install apk dependencies
RUN apk add --no-cache $PACKAGES

# Add build files
COPY Gopkg.* Makefile $REPO_PATH/
COPY .git $REPO_PATH/.git
COPY tools $REPO_PATH/tools

# Intsall go packages
RUN cd $REPO_PATH && make get_tools && make get_vendor_deps

# Add source files
COPY . $REPO_PATH

# Build app
RUN cd $REPO_PATH && make all && make install

# remove packages
RUN apk del $PACKAGES

# Set entrypoint
#ENTRYPOINT ["gaiad"]
CMD ["gaiad"]
