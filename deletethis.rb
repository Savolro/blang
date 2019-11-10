#! /usr/bin/env ruby

$keywords = {
  'if' => :KW_IF,
  'return' => :KW_RETURN,
  'while' => :KW_WHILE,
}

class Token
  attr_reader :type, :value, :line_no

  def initialize(type, value, line_no)
    @type = type
    @value = value
    @line_no = line_no
  end
end

class Lexer
  def initialize(input)
    @input = input
    @offset = 0
    @state = :START
    @tokens = []
    @line_no = 1
    @token_start_ln = 1
    @running = true
    @buffer = ""
  end

  def add
    @buffer << @curr_char
  end

  def begin_token(new_state)
    @token_start = @offset
    @token_start_ln = @line_no
    @state = new_state
  end

  def complete_token(token_type, advance = true)
    if !advance
      @offset -= 1
    end

    @tokens << Token.new(token_type, @buffer, @token_start_ln)
    @buffer = ""
    @state = :START
  end

  def dump_tokens
    puts '%2s|%2s| %-12s| %s' % ['NO', 'LN', 'TYPE', 'VALUE']
    @tokens.each_with_index do |token, index|
      puts '%2i|%2i| %-12s| %s' % [
        index, token.line_no, token.type, token.value
      ]
    end
  end

  def error
    STDERR.puts 'file.txt:%i: lexer error: unexpected symbol %s' %
      [@line_no, @curr_char]
    @running = false
  end

  def lex_all
    @offset = 0
    while @running && @offset < @input.size
      @curr_char = @input[@offset]
      # if @curr_char == "\n"; @line_no += 1; end # BAAAAD
      lex_char
      @offset += 1
    end

    @curr_char = 'EOF'
    lex_char
  end

  def lex_char
    case @state
    when :COMMENT; lex_comment
    when :IDENT; lex_ident
    when :LIT_INT; lex_lit_int
    when :LIT_STR; lex_lit_str
    when :LIT_STR_ESCAPE; lex_lit_str_escape
    when :OP_L; lex_op_l
    when :START; lex_start
    else; raise "bad"
    end
  end

  def lex_comment
    case @curr_char
    when "\n"; @line_no += 1; @state = :START
    else; # ignore
    end
  end

  def lex_ident
    if @curr_char >= 'a' && @curr_char <= 'z'
      add
    elsif @curr_char >= 'A' && @curr_char <= 'Z'
      add
    elsif @curr_char >= '0' && @curr_char <= '9'
      add
    elsif kw_type = $keywords[@buffer]
      @buffer = ""; complete_token(kw_type, false)
    else
      complete_token(:IDENT, false)
    end
  end

  def lex_lit_int
    if @curr_char >= '0' && @curr_char <= '9'
      add
    else
      complete_token(:LIT_INT, false)
    end
  end

  def lex_lit_str
    case @curr_char
    when '"'; complete_token(:LIT_STR)
    when "\\"; @state = :LIT_STR_ESCAPE
    when "\n"; add; @line_no += 1
    else; add
    end
  end

  def lex_lit_str_escape
    case @curr_char
    when '"'; @buffer << "\""
    when "n"; @buffer << "\n"
    when "t"; @buffer << "\t"
    else; error
    end
    @state = :LIT_STR
  end

  def lex_op_l
    case @curr_char
    when '='; complete_token(:OP_LE)
    else; complete_token(:OP_L, false)
    end
  end

  def lex_start
    case @curr_char
    when '#'; begin_token(:COMMENT)
    when 'a'..'z'; add; begin_token(:IDENT)
    when 'A'..'Z'; add; begin_token(:IDENT)
    when '0'..'9'; add; begin_token(:LIT_INT)
    when '+'; begin_token(:START); complete_token(:OP_PLUS)
    when '<'; begin_token(:OP_L)
    when '"'; begin_token(:LIT_STR)
    when ' ';
    when "\n"; @line_no += 1
    when 'EOF';
    else; error
    end
  end
end

#input = File.read(ARGV[0])
input = File.read('program.tm')
lexer = Lexer.new(input)
lexer.lex_all
lexer.dump_tokens
